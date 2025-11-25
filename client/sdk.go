package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	boshhttp "github.com/cloudfoundry/bosh-utils/httpclient"

	s3cli_config "github.com/cloudfoundry/bosh-s3cli/config"
)

type RecalculateV4Signature struct {
	next   http.RoundTripper
	signer *v4.Signer
	cfg    aws.Config
}

func (lt *RecalculateV4Signature) RoundTrip(req *http.Request) (*http.Response, error) {
	// Check if this is an anonymous request (no Authorization header)
	if req.Header.Get("Authorization") == "" {
		// For anonymous requests, skip signature recalculation and proceed directly
		return lt.next.RoundTrip(req)
	}

	// store for later use
	val := req.Header.Get("Accept-Encoding")

	// delete the header so the header doesn't account for in the signature
	req.Header.Del("Accept-Encoding")

	// sign with the same date
	timeString := req.Header.Get("X-Amz-Date")
	if timeString == "" {
		// If no X-Amz-Date header, this might be an unsigned request, proceed without re-signing
		req.Header.Set("Accept-Encoding", val)
		return lt.next.RoundTrip(req)
	}

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

func NewAwsS3Client(c *s3cli_config.S3Cli) (*s3.Client, error) {
	var httpClient *http.Client

	if c.SSLVerifyPeer {
		httpClient = boshhttp.CreateDefaultClient(nil)
	} else {
		httpClient = boshhttp.CreateDefaultClientInsecureSkipVerify()
	}

	options := []func(*config.LoadOptions) error{
		config.WithHTTPClient(httpClient),
	}

	if c.UseRegion() {
		options = append(options, config.WithRegion(c.Region))
	} else {
		options = append(options, config.WithRegion(s3cli_config.EmptyRegion))
	}

	if c.CredentialsSource == s3cli_config.StaticCredentialsSource {
		options = append(options, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(c.AccessKeyID, c.SecretAccessKey, ""),
		))
	}

	if c.CredentialsSource == s3cli_config.NoneCredentialsSource {
		options = append(options, config.WithCredentialsProvider(aws.AnonymousCredentials{}))
	}

	awsConfig, err := config.LoadDefaultConfig(context.TODO(), options...)
	if err != nil {
		return nil, err
	}

	if c.AssumeRoleArn != "" {
		stsClient := sts.NewFromConfig(awsConfig)
		provider := stscreds.NewAssumeRoleProvider(stsClient, c.AssumeRoleArn)
		awsConfig.Credentials = aws.NewCredentialsCache(provider)
	}

	awsConfig.RequestChecksumCalculation = aws.RequestChecksumCalculationUnset
	awsConfig.HTTPClient = &http.Client{Transport: &RecalculateV4Signature{http.DefaultTransport, v4.NewSigner(), awsConfig}}

	s3Client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		o.UsePathStyle = !c.HostStyle
		if c.S3Endpoint() != "" {
			endpoint := c.S3Endpoint()
			// AWS SDK v2 requires full URI with protocol
			if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
				if c.UseSSL {
					endpoint = "https://" + endpoint
				} else {
					endpoint = "http://" + endpoint
				}
			}
			o.BaseEndpoint = aws.String(endpoint)
		}
	})

	return s3Client, nil
}
