package integration_test

import (
	"os"
	"strings"

	"github.com/cloudfoundry/bosh-s3cli/config"
	"github.com/cloudfoundry/bosh-s3cli/integration"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing gets against a public AWS S3 bucket", func() {
	Context("with PUBLIC READ ONLY (no creds) configuration", func() {
		It("can successfully get a publicly readable file", func() {
			accessKeyID := os.Getenv("ACCESS_KEY_ID")
			Expect(accessKeyID).ToNot(BeEmpty(), "ACCESS_KEY_ID must be set")

			secretAccessKey := os.Getenv("SECRET_ACCESS_KEY")
			Expect(secretAccessKey).ToNot(BeEmpty(), "SECRET_ACCESS_KEY must be set")

			bucketName := os.Getenv("BUCKET_NAME")
			Expect(bucketName).ToNot(BeEmpty(), "BUCKET_NAME must be set")

			region := os.Getenv("REGION")
			Expect(region).ToNot(BeEmpty(), "REGION must be set")

			s3Filename := integration.GenerateRandomString()
			s3FileContents := integration.GenerateRandomString()

			awsSession, err := session.NewSession(
				aws.NewConfig().
					WithCredentials(credentials.NewStaticCredentials(accessKeyID, secretAccessKey, "")).
					WithRegion(region),
			)
			Expect(err).ToNot(HaveOccurred())
			s3Client := s3.New(awsSession)

			_, err = s3Client.PutObject(&s3.PutObjectInput{
				Body:   strings.NewReader(s3FileContents),
				Bucket: &bucketName,
				Key:    &s3Filename,
				ACL:    aws.String(s3.ObjectCannedACLPublicRead),
			})
			Expect(err).ToNot(HaveOccurred())

			cfg := &config.S3Cli{
				BucketName: bucketName,
				Region:     region,
			}

			configPath := integration.MakeConfigFile(cfg)
			defer os.Remove(configPath) //nolint:errcheck

			s3CLISession, err := integration.RunS3CLI(s3CLIPath, configPath, "get", s3Filename, "public-file")
			Expect(err).ToNot(HaveOccurred())

			defer os.Remove("public-file") //nolint:errcheck
			Expect(s3CLISession.ExitCode()).To(BeZero())

			gottenBytes, err := os.ReadFile("public-file")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(gottenBytes)).To(Equal(s3FileContents))

			s3CLISession, err = integration.RunS3CLI(s3CLIPath, configPath, "exists", s3Filename)
			Expect(err).ToNot(HaveOccurred())
			Expect(s3CLISession.ExitCode()).To(BeZero())
			Expect(s3CLISession.Err.Contents()).To(MatchRegexp("File '.*' exists in bucket '.*'"))
		})
	})
})
