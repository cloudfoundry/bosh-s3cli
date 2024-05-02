package integration_test

import (
	"bytes"
	"os"

	"github.com/cloudfoundry/bosh-s3cli/config"
	"github.com/cloudfoundry/bosh-s3cli/integration"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing for working signed URLs all Swift/OpenStack regions", func() {

	Context("with GENERAL OpenStack/Swift (static creds) configurations", func() {

		var configPath string
		var contentFile string
		var defaultConfig config.S3Cli

		accessKeyID := os.Getenv("ACCESS_KEY_ID")
		secretAccessKey := os.Getenv("SECRET_ACCESS_KEY")
		bucketName := os.Getenv("BUCKET_NAME")
		region := os.Getenv("REGION")
		swiftHost := os.Getenv("SWIFT_HOST")
		swiftTempURLKey := os.Getenv("SWIFT_TEMPURL_KEY")
		swiftAuthAccount := os.Getenv("SWIFT_AUTH_ACCOUNT")
		s3CLIPath := os.Getenv("S3_CLI_PATH")

		BeforeEach(func() {})

		BeforeEach(func() {
			if os.Getenv("SWIFT_AUTH_ACCOUNT") == "" {
				Skip("Skipping because swift blobstore isn't available")
			}
			Expect(accessKeyID).ToNot(BeEmpty(), "ACCESS_KEY_ID must be set")
			Expect(secretAccessKey).ToNot(BeEmpty(), "SECRET_ACCESS_KEY must be set")
			Expect(bucketName).ToNot(BeEmpty(), "BUCKET_NAME must be set")
			Expect(region).ToNot(BeEmpty(), "REGION must be set")
			Expect(swiftTempURLKey).ToNot(BeEmpty(), "SWIFT_TEMPURL_KEY must be set")
			Expect(swiftAuthAccount).ToNot(BeEmpty(), "SWIFT_AUTH_ACCOUNT must be set")
			defaultConfig = config.S3Cli{
				AccessKeyID:      accessKeyID,
				SecretAccessKey:  secretAccessKey,
				BucketName:       bucketName,
				Host:             swiftHost,
				SwiftAuthAccount: swiftAuthAccount,
			}
			configPath = integration.MakeConfigFile(&defaultConfig)
			contentFile = integration.MakeContentFile("foo")
		})

		AfterEach(func() {
			defer func() { _ = os.Remove(configPath) }()
			defer func() { _ = os.Remove(contentFile) }()
		})

		Describe("Invoking `sign`", func() {

			It("returns 0 for an existing blob", func() {

				cliSession, err := integration.RunS3CLI(s3CLIPath, configPath, "sign", "some-blob", "get", "60s")
				Expect(err).ToNot(HaveOccurred())
				Expect(cliSession.ExitCode()).To(BeZero())

				getUrl := bytes.NewBuffer(cliSession.Out.Contents()).String()
				Expect(getUrl).To(MatchRegexp("https://" + swiftHost + ".*?" + "/some-blob"))
				cliSession, err = integration.RunS3CLI(s3CLIPath, configPath, "sign", "some-blob", "put", "60s")
				Expect(err).ToNot(HaveOccurred())

				putUrl := bytes.NewBuffer(cliSession.Out.Contents()).String()
				Expect(putUrl).To(MatchRegexp("https://" + swiftHost + ".*?" + "/some-blob"))
			})
		})
	})
})
