package integration

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/cloudfoundry/bosh-s3cli/config"
	boshhttp "github.com/cloudfoundry/bosh-utils/httpclient"
)

type RecalculateV4Signature struct {
	next   http.RoundTripper
	signer *v4.Signer
	cfg    aws.Config
}

func (lt *RecalculateV4Signature) RoundTrip(req *http.Request) (*http.Response, error) {
	// store for later use
	val := req.Header.Get("Accept-Encoding")

	// delete the header so the header doesn't account for in the signature
	req.Header.Del("Accept-Encoding")

	// sign with the same date
	timeString := req.Header.Get("X-Amz-Date")
	timeDate, err := time.Parse("20060102T150405Z", timeString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse X-Amz-Date header: %w", err)
	}

	creds, err := lt.cfg.Credentials.Retrieve(req.Context())
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve credentials: %w", err)
	}

	err = lt.signer.SignHTTP(req.Context(), creds, req, v4.GetPayloadHash(req.Context()), "s3", lt.cfg.Region, timeDate)
	if err != nil {
		return nil, err
	}
	// Reset Accept-Encoding if desired
	req.Header.Set("Accept-Encoding", val)

	// follows up the original round tripper
	return lt.next.RoundTrip(req)
}

// CreateUploadPartTracker creates an Initialize middleware that tracks upload parts
func CreateUploadPartTracker() middleware.InitializeMiddleware {
	var partCounter int64

	return middleware.InitializeMiddlewareFunc("UploadPartTracker", func(
		ctx context.Context, in middleware.InitializeInput, next middleware.InitializeHandler,
	) (
		out middleware.InitializeOutput, metadata middleware.Metadata, err error,
	) {
		// Type switch to check if the input is s3.UploadPartInput
		injectFailure := false
		switch in.Parameters.(type) {
		case *s3.UploadPartInput:
			// Increment counter and mark even-numbered parts for failure
			count := atomic.AddInt64(&partCounter, 1)
			if count%2 == 0 {
				injectFailure = true
			}
		}

		// Store the injectFailure flag in the context for this request
		ctx = context.WithValue(ctx, failureInjectionKey{}, injectFailure)

		// Continue to next middleware
		return next.HandleInitialize(ctx, in)
	})
}

// Context key for failure injection flag
type failureInjectionKey struct{}

// CreateSHACorruptionMiddleware creates a Finalize middleware that corrupts headers
func CreateSHACorruptionMiddleware() middleware.FinalizeMiddleware {
	return middleware.FinalizeMiddlewareFunc("SHACorruptionMiddleware", func(
		ctx context.Context, in middleware.FinalizeInput, next middleware.FinalizeHandler,
	) (middleware.FinalizeOutput, middleware.Metadata, error) {
		// Check if we should inject a failure based on context value
		if inject, ok := ctx.Value(failureInjectionKey{}).(bool); ok && inject {
			if req, ok := in.Request.(*smithyhttp.Request); ok {
				// Corrupt the SHA256 header to cause upload failure
				req.Header.Set("X-Amz-Content-Sha256", "000")
			}
		}
		return next.HandleFinalize(ctx, in)
	})
}

// CreateS3ClientWithFailureInjection creates an S3 client with failure injection middleware
func CreateS3ClientWithFailureInjection(s3Config *config.S3Cli) (*s3.Client, error) {
	// Create HTTP client based on SSL verification settings
	var httpClient *http.Client
	if s3Config.SSLVerifyPeer {
		httpClient = boshhttp.CreateDefaultClient(nil)
	} else {
		httpClient = boshhttp.CreateDefaultClientInsecureSkipVerify()
	}

	// Set up AWS config options
	options := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithHTTPClient(httpClient),
	}

	if s3Config.UseRegion() {
		options = append(options, awsconfig.WithRegion(s3Config.Region))
	} else {
		options = append(options, awsconfig.WithRegion(config.EmptyRegion))
	}

	if s3Config.CredentialsSource == config.StaticCredentialsSource {
		options = append(options, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(s3Config.AccessKeyID, s3Config.SecretAccessKey, ""),
		))
	}

	if s3Config.CredentialsSource == config.NoneCredentialsSource {
		options = append(options, awsconfig.WithCredentialsProvider(aws.AnonymousCredentials{}))
	}

	// Load AWS config
	awsConfig, err := awsconfig.LoadDefaultConfig(context.TODO(), options...)
	if err != nil {
		return nil, err
	}

	// Handle STS assume role if configured
	if s3Config.AssumeRoleArn != "" {
		stsClient := sts.NewFromConfig(awsConfig)
		provider := stscreds.NewAssumeRoleProvider(stsClient, s3Config.AssumeRoleArn)
		awsConfig.Credentials = aws.NewCredentialsCache(provider)
	}

	// Create failure injection middlewares
	trackingMiddleware := CreateUploadPartTracker()
	corruptionMiddleware := CreateSHACorruptionMiddleware()

	// Create S3 client with custom middleware and options
	s3Client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		o.UsePathStyle = !s3Config.HostStyle
		if s3Config.S3Endpoint() != "" {
			o.BaseEndpoint = aws.String(s3Config.S3Endpoint())
		}

		// Add the failure injection middlewares
		o.APIOptions = append(o.APIOptions, func(stack *middleware.Stack) error {
			// Add initialize middleware to track UploadPart operations
			if err := stack.Initialize.Add(trackingMiddleware, middleware.Before); err != nil {
				return err
			}
			// Add finalize middleware to corrupt headers after signing
			return stack.Finalize.Add(corruptionMiddleware, middleware.After)
		})
	})

	return s3Client, nil
}

// S3TracingMiddleware captures S3 operation names for testing
type S3TracingMiddleware struct {
	calls *[]string
}

// CreateS3TracingMiddleware creates a middleware that tracks S3 operation calls
func CreateS3TracingMiddleware(calls *[]string) *S3TracingMiddleware {
	return &S3TracingMiddleware{calls: calls}
}

// ID returns the middleware identifier
func (m *S3TracingMiddleware) ID() string {
	return "S3TracingMiddleware"
}

// HandleInitialize implements the InitializeMiddleware interface
func (m *S3TracingMiddleware) HandleInitialize(
	ctx context.Context, in middleware.InitializeInput, next middleware.InitializeHandler,
) (middleware.InitializeOutput, middleware.Metadata, error) {
	// Extract operation name from the middleware metadata
	if operationName := middleware.GetStackValue(ctx, "operation"); operationName != nil {
		if opName, ok := operationName.(string); ok {
			*m.calls = append(*m.calls, opName)
		}
	}

	// Try to determine operation from the input type as fallback
	switch in.Parameters.(type) {
	case *s3.CreateMultipartUploadInput:
		*m.calls = append(*m.calls, "CreateMultipart")
	case *s3.UploadPartInput:
		*m.calls = append(*m.calls, "UploadPart")
	case *s3.CompleteMultipartUploadInput:
		*m.calls = append(*m.calls, "CompleteMultipart")
	case *s3.PutObjectInput:
		*m.calls = append(*m.calls, "PutObject")
	case *s3.GetObjectInput:
		*m.calls = append(*m.calls, "GetObject")
	case *s3.DeleteObjectInput:
		*m.calls = append(*m.calls, "DeleteObject")
	case *s3.HeadObjectInput:
		*m.calls = append(*m.calls, "HeadObject")
	}

	return next.HandleInitialize(ctx, in)
}

// CreateS3ClientWithTracing creates a new S3 client with tracing middleware
func CreateS3ClientWithTracing(baseClient *s3.Client, tracingMiddleware *S3TracingMiddleware) *s3.Client {
	// Create a wrapper that captures calls and delegates to the base client
	// Since AWS SDK v2 makes it difficult to extract config from existing clients,
	// we'll use a different approach: modify the traceS3 function to work differently

	// For the tracing functionality, we'll need to intercept at a higher level
	// The current implementation will track operations through the middleware
	// that inspects the input parameters

	return baseClient
}

// CreateTracingS3Client creates an S3 client with tracing middleware from config
func CreateTracingS3Client(s3Config *config.S3Cli, calls *[]string) (*s3.Client, error) {
	// Create HTTP client based on SSL verification settings
	var httpClient *http.Client
	if s3Config.SSLVerifyPeer {
		httpClient = boshhttp.CreateDefaultClient(nil)
	} else {
		httpClient = boshhttp.CreateDefaultClientInsecureSkipVerify()
	}

	// Set up AWS config options
	options := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithHTTPClient(httpClient),
	}

	if s3Config.UseRegion() {
		options = append(options, awsconfig.WithRegion(s3Config.Region))
	} else {
		options = append(options, awsconfig.WithRegion(config.EmptyRegion))
	}

	if s3Config.CredentialsSource == config.StaticCredentialsSource {
		options = append(options, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(s3Config.AccessKeyID, s3Config.SecretAccessKey, ""),
		))
	}

	if s3Config.CredentialsSource == config.NoneCredentialsSource {
		options = append(options, awsconfig.WithCredentialsProvider(aws.AnonymousCredentials{}))
	}

	// Load AWS config
	awsConfig, err := awsconfig.LoadDefaultConfig(context.TODO(), options...)
	if err != nil {
		return nil, err
	}

	// Handle STS assume role if configured
	if s3Config.AssumeRoleArn != "" {
		stsClient := sts.NewFromConfig(awsConfig)
		provider := stscreds.NewAssumeRoleProvider(stsClient, s3Config.AssumeRoleArn)
		awsConfig.Credentials = aws.NewCredentialsCache(provider)
	}

	awsConfig.RequestChecksumCalculation = aws.RequestChecksumCalculationUnset
	awsConfig.HTTPClient = &http.Client{Transport: &RecalculateV4Signature{http.DefaultTransport, v4.NewSigner(), awsConfig}}

	// Create tracing middleware
	tracingMiddleware := CreateS3TracingMiddleware(calls)

	// Create S3 client with tracing middleware
	s3Client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		o.UsePathStyle = !s3Config.HostStyle
		if s3Config.S3Endpoint() != "" {
			o.BaseEndpoint = aws.String(s3Config.S3Endpoint())
		}

		// Add the tracing middleware
		o.APIOptions = append(o.APIOptions, func(stack *middleware.Stack) error {
			return stack.Initialize.Add(tracingMiddleware, middleware.Before)
		})
	})

	return s3Client, nil
}
