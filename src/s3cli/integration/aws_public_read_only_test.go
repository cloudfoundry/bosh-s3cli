package integration_test

import (
	"io/ioutil"
	"os"
	"s3cli/config"
	"s3cli/integration"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing gets against a public AWS S3 bucket", func() {
	Context("with PUBLIC READ ONLY (no creds) configuration", func() {
		It("can successfully get a publicly readable file", func() {
			s3CLIPath := os.Getenv("S3_CLI_PATH")
			Expect(s3CLIPath).ToNot(BeEmpty(), "S3_CLI_PATH must be set")

			bucketName := os.Getenv("BUCKET_NAME")
			Expect(bucketName).ToNot(BeEmpty(), "BUCKET_NAME must be set")

			region := os.Getenv("REGION")
			Expect(region).ToNot(BeEmpty(), "REGION must be set")

			s3PublicFile := os.Getenv("S3_PUBLIC_FILE")
			Expect(s3PublicFile).ToNot(BeEmpty(), "S3_PUBLIC_FILE must be set")

			s3FileContents := os.Getenv("S3_FILE_CONTENTS")
			Expect(s3FileContents).ToNot(BeEmpty(), "S3_FILE_CONTENTS must be set")

			cfg := &config.S3Cli{
				BucketName: bucketName,
				Region:     region,
			}

			configPath := integration.MakeConfigFile(cfg)
			defer func() { _ = os.Remove(configPath) }()

			s3CLISession, err := integration.RunS3CLI(s3CLIPath, configPath, "get", s3PublicFile, "public-file")
			Expect(err).ToNot(HaveOccurred())

			defer func() { _ = os.Remove("public-file") }()
			Expect(s3CLISession.ExitCode()).To(BeZero())

			gottenBytes, err := ioutil.ReadFile("public-file")
			Expect(err).ToNot(HaveOccurred())
			Expect(gottenBytes).To(Equal(s3FileContents))

			s3CLISession, err = integration.RunS3CLI(s3CLIPath, configPath, "exists", s3PublicFile)
			Expect(err).ToNot(HaveOccurred())
			Expect(s3CLISession.ExitCode()).To(BeZero())
			Expect(s3CLISession.Out.Contents()).To(MatchRegexp("File '.*' exists in bucket '.*'"))
		})
	})
})
