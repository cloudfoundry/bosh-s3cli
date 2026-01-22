package client

import (
	"net/http"
	"strings"

	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	boshhttp "github.com/cloudfoundry/bosh-utils/httpclient"

	s3cli_config "github.com/cloudfoundry/bosh-s3cli/config"
)

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
		if c.Debug {
			o.ClientLogMode = aws.LogResponse | aws.LogRequest
		}
	})

	return s3Client, nil
}
