package integration_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"s3cli/config"
	"s3cli/integration"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing in any AWS region isolated from the US standard regions (i.e., cn-north-1)", func() {
	Context("with ISOLATED REGION (static creds) configurations", func() {
		It("fails with a config that specifies a valid region but invalid host", func() {
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

			cfg := &config.S3Cli{
				SignatureVersion:  4,
				CredentialsSource: "static",
				AccessKeyID:       accessKeyID,
				SecretAccessKey:   secretAccessKey,
				BucketName:        bucketName,
				Region:            region,
				Host:              "s3-external-1.amazonaws.com",
			}
			s3Filename, err := integration.GenerateRandomString()
			Expect(err).ToNot(HaveOccurred())

			configPath := integration.MakeConfigFile(cfg)
			defer func() { _ = os.Remove(configPath) }()

			err = ioutil.WriteFile(s3Filename, []byte("test"), 0644)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = os.Remove(s3Filename) }()

			s3CLISession, err := integration.RunS3CLI(s3CLIPath, configPath, "put", s3Filename, fmt.Sprintf("s3://%s/", bucketName))
			Expect(err).ToNot(HaveOccurred())
			Expect(s3CLISession.ExitCode()).ToNot(BeZero())
			Expect(s3CLISession.Out.Contents()).To(ContainSubstring("InvalidAccessKeyId"))

			s3CLISession, err = integration.RunS3CLI(s3CLIPath, configPath, "delete", s3Filename)
			Expect(err).ToNot(HaveOccurred())
			Expect(s3CLISession.ExitCode()).ToNot(BeZero())
			Expect(s3CLISession.Out.Contents()).To(ContainSubstring("InvalidAccessKeyId"))
		})
	})
})
