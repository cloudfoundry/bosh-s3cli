package integration_test

import (
	"os"
	"s3cli/config"
	"s3cli/integration"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing in any non-AWS, S3 compatible storage service", func() {
	Context("with S3 COMPATIBLE (static creds) configurations", func() {
		s3CLIPath := os.Getenv("S3_CLI_PATH")
		accessKeyID := os.Getenv("ACCESS_KEY_ID")
		secretAccessKey := os.Getenv("SECRET_ACCESS_KEY")
		bucketName := os.Getenv("BUCKET_NAME")
		s3Host := os.Getenv("S3_HOST")
		s3PortString := os.Getenv("S3_PORT")
		s3Port, atoiErr := strconv.Atoi(s3PortString)

		BeforeEach(func() {
			Expect(s3CLIPath).ToNot(BeEmpty(), "S3_CLI_PATH must be set")
			Expect(accessKeyID).ToNot(BeEmpty(), "ACCESS_KEY_ID must be set")
			Expect(secretAccessKey).ToNot(BeEmpty(), "SECRET_ACCESS_KEY must be set")
			Expect(bucketName).ToNot(BeEmpty(), "BUCKET_NAME must be set")
			Expect(s3Host).ToNot(BeEmpty(), "S3_HOST must be set")
			Expect(s3PortString).ToNot(BeEmpty(), "S3_PORT must be set")
			Expect(atoiErr).ToNot(HaveOccurred())
		})

		configurations := []TableEntry{
			Entry("with the minimal configuration", &config.S3Cli{
				AccessKeyID:     accessKeyID,
				SecretAccessKey: secretAccessKey,
				BucketName:      bucketName,
				Host:            s3Host,
			}),
			Entry("with region specified", &config.S3Cli{
				AccessKeyID:     accessKeyID,
				SecretAccessKey: secretAccessKey,
				BucketName:      bucketName,
				Host:            s3Host,
				Region:          "invalid-region",
			}),
			Entry("with use_ssl set to false", &config.S3Cli{
				AccessKeyID:     accessKeyID,
				SecretAccessKey: secretAccessKey,
				BucketName:      bucketName,
				Host:            s3Host,
				UseSSL:          false,
			}),
			Entry("with the maximal configuration", &config.S3Cli{
				SignatureVersion:  2,
				CredentialsSource: "static",
				AccessKeyID:       accessKeyID,
				SecretAccessKey:   secretAccessKey,
				BucketName:        bucketName,
				Host:              s3Host,
				Port:              s3Port,
				UseSSL:            true,
				SSLVerifyPeer:     true,
				Region:            "invalid-region",
			}),
		}

		DescribeTable("Blobstore lifecycle works",
			func(cfg *config.S3Cli) { integration.AssertLifecycleWorks(s3CLIPath, cfg) },
			configurations...,
		)
		DescribeTable("Invoking `s3cli get` on a non-existent-key fails",
			func(cfg *config.S3Cli) { integration.AssertGetNonexistentFails(s3CLIPath, cfg) },
			configurations...,
		)
		DescribeTable("Invoking `s3cli delete` on a non-existent-key does not fail",
			func(cfg *config.S3Cli) { integration.AssertDeleteNonexistentWorks(s3CLIPath, cfg) },
			configurations...,
		)

		It("fails with a config that specifies signature version 4", func() {
			cfg := &config.S3Cli{
				SignatureVersion: 4,
				AccessKeyID:      accessKeyID,
				SecretAccessKey:  secretAccessKey,
				BucketName:       bucketName,
				Host:             s3Host,
			}
			s3Filename, err := integration.GenerateRandomString()
			Expect(err).ToNot(HaveOccurred())

			configPath := integration.MakeConfigFile(cfg)
			defer func() { _ = os.Remove(configPath) }()

			contentFile := integration.MakeContentFile("test")
			defer func() { _ = os.Remove(contentFile) }()

			s3CLISession, err := integration.RunS3CLI(s3CLIPath, configPath, "put", contentFile, s3Filename)
			Expect(err).ToNot(HaveOccurred())
			Expect(s3CLISession.ExitCode()).ToNot(BeZero())

			s3CLISession, err = integration.RunS3CLI(s3CLIPath, configPath, "delete", s3Filename)
			Expect(err).ToNot(HaveOccurred())
			Expect(s3CLISession.ExitCode()).ToNot(BeZero())
		})
	})
})
