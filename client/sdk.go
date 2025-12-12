package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	boshhttp "github.com/cloudfoundry/bosh-utils/httpclient"

	s3cli_config "github.com/cloudfoundry/bosh-s3cli/config"
)

const acceptEncodingHeader = "Accept-Encoding"

type acceptEncodingKey struct{}

func GetAcceptEncodingKey(ctx context.Context) (v string) {
	v, _ = middleware.GetStackValue(ctx, acceptEncodingKey{}).(string)
	return v
}

func SetAcceptEncodingKey(ctx context.Context, value string) context.Context {
	return middleware.WithStackValue(ctx, acceptEncodingKey{}, value)
}

var dropAcceptEncodingHeader = middleware.FinalizeMiddlewareFunc("DropAcceptEncodingHeader",
	func(ctx context.Context, in middleware.FinalizeInput, next middleware.FinalizeHandler) (out middleware.FinalizeOutput, metadata middleware.Metadata, err error) {
		req, ok := in.Request.(*smithyhttp.Request)
		if !ok {
			return out, metadata, &v4.SigningError{Err: fmt.Errorf("unexpected request middleware type %T", in.Request)}
		}

		ae := req.Header.Get(acceptEncodingHeader)
		ctx = SetAcceptEncodingKey(ctx, ae)
		req.Header.Del(acceptEncodingHeader)
		in.Request = req

		return next.HandleFinalize(ctx, in)
	},
)

var replaceAcceptEncodingHeader = middleware.FinalizeMiddlewareFunc("ReplaceAcceptEncodingHeader",
	func(ctx context.Context, in middleware.FinalizeInput, next middleware.FinalizeHandler) (out middleware.FinalizeOutput, metadata middleware.Metadata, err error) {
		req, ok := in.Request.(*smithyhttp.Request)
		if !ok {
			return out, metadata, &v4.SigningError{Err: fmt.Errorf("unexpected request middleware type %T", in.Request)}
		}

		ae := GetAcceptEncodingKey(ctx)
		req.Header.Set(acceptEncodingHeader, ae)
		in.Request = req

		return next.HandleFinalize(ctx, in)
	},
)

func NewAwsS3Client(c *s3cli_config.S3Cli, useFixSigningMiddleware bool) (*s3.Client, error) {
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
		if useFixSigningMiddleware {
			o.APIOptions = append(o.APIOptions, func(stack *middleware.Stack) error {
				if err := stack.Finalize.Insert(dropAcceptEncodingHeader, "Signing", middleware.Before); err != nil {
					return err
				}

				if err := stack.Finalize.Insert(replaceAcceptEncodingHeader, "Signing", middleware.After); err != nil {
					return err
				}
				return nil
			})
		}
	})

	return s3Client, nil
}
