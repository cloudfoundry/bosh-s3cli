package integration

import (
	"context"
	"net/http"
	"sync/atomic"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go/middleware"
	"github.com/cloudfoundry/bosh-s3cli/config"
	boshhttp "github.com/cloudfoundry/bosh-utils/httpclient"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// injectFailureKey is the context key for marking requests that should fail
	injectFailureKey contextKey = "inject-failure"
)

// CreateFailureInjectionMiddleware creates a middleware that will inject SHA mismatches on even-numbered upload parts
func CreateFailureInjectionMiddleware() middleware.InitializeMiddleware {
	var partCounter int64

	return middleware.InitializeMiddlewareFunc("FailureInjectionMiddleware", func(
		ctx context.Context, in middleware.InitializeInput, next middleware.InitializeHandler,
	) (
		out middleware.InitializeOutput, metadata middleware.Metadata, err error,
	) {
		// Type switch to check if the input is s3.UploadPartInput
		switch in.Parameters.(type) {
		case *s3.UploadPartInput:
			// Increment counter and fail every other request (even numbers)
			count := atomic.AddInt64(&partCounter, 1)
			if count%2 == 0 { // Fail even-numbered parts (2nd, 4th, etc.)
				// Add a marker to the context that the Finalize step can check
				ctx = context.WithValue(ctx, injectFailureKey, true)
			}
		}

		// Continue to next middleware
		return next.HandleInitialize(ctx, in)
	})
}

// CreateSHACorruptionMiddleware creates a finalize middleware that corrupts SHA headers
func CreateSHACorruptionMiddleware() middleware.FinalizeMiddleware {
	return middleware.FinalizeMiddlewareFunc("SHACorruptionMiddleware", func(
		ctx context.Context, in middleware.FinalizeInput, next middleware.FinalizeHandler,
	) (middleware.FinalizeOutput, middleware.Metadata, error) {
		// Check if we should inject a failure based on context
		if shouldInject, ok := ctx.Value(injectFailureKey).(bool); ok && shouldInject {
			// Check if this is an HTTP request
			if req, ok := in.Request.(*http.Request); ok {
				// Modify the request to have an invalid SHA256 checksum
				req.Header.Set("X-Amz-Content-Sha256", "invalid-checksum-to-cause-failure")
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
	initializeMiddleware := CreateFailureInjectionMiddleware()
	finalizeMiddleware := CreateSHACorruptionMiddleware()

	// Create S3 client with custom middleware and options
	s3Client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		o.UsePathStyle = !s3Config.HostStyle
		if s3Config.S3Endpoint() != "" {
			o.BaseEndpoint = aws.String(s3Config.S3Endpoint())
		}

		// Add the failure injection middlewares
		o.APIOptions = append(o.APIOptions, func(stack *middleware.Stack) error {
			// Add initialize middleware to track UploadPart operations
			if err := stack.Initialize.Add(initializeMiddleware, middleware.Before); err != nil {
				return err
			}
			// Add finalize middleware to corrupt SHA headers when needed
			return stack.Finalize.Add(finalizeMiddleware, middleware.After)
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
