package integration_test

import (
	"io/ioutil"
	"os"
	"s3cli/config"
	"s3cli/integration"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing gets against a public AWS S3 bucket", func() {
	Context("with PUBLIC READ ONLY (no creds) configuration", func() {
		It("can successfully get a publicly readable file", func() {
			s3CLIPath := os.Getenv("S3_CLI_PATH")
			Expect(s3CLIPath).ToNot(BeEmpty(), "S3_CLI_PATH must be set")

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

			s3Client := s3.New(session.New(&aws.Config{
				Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
				Region:      aws.String(region),
			}))

			_, err := s3Client.PutObject(&s3.PutObjectInput{
				Body:   strings.NewReader(s3FileContents),
				Bucket: &bucketName,
				Key:    &s3Filename,
			})
			Expect(err).ToNot(HaveOccurred())

			cfg := &config.S3Cli{
				BucketName: bucketName,
				Region:     region,
			}

			configPath := integration.MakeConfigFile(cfg)
			defer func() { _ = os.Remove(configPath) }()

			s3CLISession, err := integration.RunS3CLI(s3CLIPath, configPath, "get", s3Filename, "public-file")
			Expect(err).ToNot(HaveOccurred())

			defer func() { _ = os.Remove("public-file") }()
			Expect(s3CLISession.ExitCode()).To(BeZero())

			gottenBytes, err := ioutil.ReadFile("public-file")
			Expect(err).ToNot(HaveOccurred())
			Expect(gottenBytes).To(Equal(s3FileContents))

			s3CLISession, err = integration.RunS3CLI(s3CLIPath, configPath, "exists", s3Filename)
			Expect(err).ToNot(HaveOccurred())
			Expect(s3CLISession.ExitCode()).To(BeZero())
			Expect(s3CLISession.Out.Contents()).To(MatchRegexp("File '.*' exists in bucket '.*'"))
		})
	})
})
