package client_test

import (
	"bytes"
	"time"

	"github.com/cloudfoundry/bosh-s3cli/client"
	"github.com/cloudfoundry/bosh-s3cli/config"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SwiftBlobstore", func() {
	var blobstore client.SwiftBlobstore
	var s3CliConfig *config.S3Cli

	BeforeEach(func() {
		configBytes := []byte(`{
			"access_key_id":      "id",
			"secret_access_key":  "key",
			"bucket_name":        "some-bucket",
			"host": 			  "host-name",
			"swift_auth_account": "swift_account",
			"swift_temp_url_key": "temp_key"
		}`)

		configReader := bytes.NewReader(configBytes)
		s3CliConfigCp, _ := config.NewFromReader(configReader)
		s3CliConfig = &s3CliConfigCp
		blobstore = client.NewSwiftClient(s3CliConfig)
	})

	Describe("#Sign", func() {
		It("returns a signed URL for GET action", func() {
			objectID := "test-object-id"
			action := "GET"
			expiration := time.Duration(100) * time.Second

			url, err := blobstore.Sign(objectID, action, expiration)
			Expect(err).NotTo(HaveOccurred())
			Expect(url).To(HavePrefix("https://host-name"))
			Expect(url).To(ContainSubstring("temp_url_sig"))
			Expect(url).To(ContainSubstring("temp_url_expires"))
		})

		It("returns a signed URL for PUT action", func() {
			objectID := "test-object-id"
			action := "PUT"
			expiration := time.Duration(100) * time.Second

			url, err := blobstore.Sign(objectID, action, expiration)
			Expect(err).NotTo(HaveOccurred())
			Expect(url).To(HavePrefix("https://host-name"))
			Expect(url).To(ContainSubstring("temp_url_sig"))
			Expect(url).To(ContainSubstring("temp_url_expires"))
		})

		It("returns an error for unsupported action", func() {
			objectID := "test-object-id"
			action := "UNSUPPORTED_ACTION"
			expiration := time.Duration(100) * time.Second

			_, err := blobstore.Sign(objectID, action, expiration)
			Expect(err).To(HaveOccurred())
		})
	})
})
