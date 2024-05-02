package client_test

import (
	"bytes"

	"github.com/cloudfoundry/bosh-s3cli/config"

	"github.com/cloudfoundry/bosh-s3cli/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {

	Describe("SignURLProvider", func() {
		It("returns a SwiftBlobstore client", func() {
			configBytes := []byte(`{
				"access_key_id":      "id",
				"secret_access_key":  "key",
				"bucket_name":        "some-bucket",
				"swift_auth_account": "swift_account"
			}`)

			configReader := bytes.NewReader(configBytes)
			s3CliConfig, _ := config.NewFromReader(configReader)
			s3Client, _ := client.NewSDK(s3CliConfig)
			blobstoreClient, _ := client.New(s3Client, &s3CliConfig)

			signURLProvider, err := client.NewSignURLProvider(blobstoreClient, &s3CliConfig)
			Expect(err).ToNot(HaveOccurred())
			Expect(signURLProvider).To(BeAssignableToTypeOf(&client.SwiftBlobstore{}))
		})
		It("returns a S3Blobstore client", func() {
			configBytes := []byte(`{
				"access_key_id":      "id",
				"secret_access_key":  "key",
				"bucket_name":        "some-bucket"
			}`)

			configReader := bytes.NewReader(configBytes)
			s3CliConfig, _ := config.NewFromReader(configReader)
			s3Client, _ := client.NewSDK(s3CliConfig)
			blobstoreClient, _ := client.New(s3Client, &s3CliConfig)

			signURLProvider, err := client.NewSignURLProvider(blobstoreClient, &s3CliConfig)
			Expect(err).ToNot(HaveOccurred())
			Expect(signURLProvider).To(BeAssignableToTypeOf(&client.S3Blobstore{}))
		})
	})
})
