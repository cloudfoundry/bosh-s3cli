package client_test

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/cloudfoundry/bosh-s3cli/client"
	"github.com/cloudfoundry/bosh-s3cli/config"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("S3CompatibleClient", func() {
	var blobstoreClient client.S3CompatibleClient
	var s3Config *config.S3Cli

	Describe("Sign()", func() {
		var objectId = "test-object-id"
		var expiration = time.Duration(100) * time.Second
		var action string
		var urlRegexp string

		Context("when SwiftAuthAccount is empty", func() {
			BeforeEach(func() {
				s3Config = &config.S3Cli{
					AccessKeyID:     "id",
					SecretAccessKey: "key",
					BucketName:      "some-bucket",
					Host:            "host-name",
				}
				awsCfg := aws.Config{
					Region: "us-west-2",
					Credentials: credentials.NewStaticCredentialsProvider(
						s3Config.AccessKeyID,
						s3Config.SecretAccessKey,
						"",
					),
				}

				s3Client := s3.NewFromConfig(awsCfg)

				blobstoreClient = client.New(s3Client, s3Config)

				urlRegexp = `https://some-bucket.s3.us-west-2.amazonaws.com/test-object-id` +
					`\?X-Amz-Algorithm=AWS4-HMAC-SHA256` +
					`&X-Amz-Credential=id%2F([0-9]+)%2Fus-west-2%2Fs3%2Faws4_request` +
					`&X-Amz-Date=([0-9]+)T([0-9]+)Z` +
					`&X-Amz-Expires=100` +
					`&X-Amz-SignedHeaders=host` +
					`&x-id=[A-Za-z]+` +
					`&X-Amz-Signature=([a-f0-9]+)`
			})

			Context("when the action is GET", func() {
				BeforeEach(func() {
					action = "GET"
				})

				It("returns a signed URL", func() {
					url, err := blobstoreClient.Sign(objectId, action, expiration)
					Expect(err).NotTo(HaveOccurred())

					Expect(url).To(MatchRegexp(urlRegexp))
				})
			})

			Context("when the action is PUT", func() {
				BeforeEach(func() {
					action = "PUT"
				})

				It("returns a signed URL", func() {
					url, err := blobstoreClient.Sign(objectId, action, expiration)
					Expect(err).NotTo(HaveOccurred())

					Expect(url).To(MatchRegexp(urlRegexp))
				})
			})

			Context("when the action is neither GET nor PUT", func() {
				BeforeEach(func() {
					action = "UNSUPPORTED_ACTION"
				})

				It("returns an error", func() {
					_, err := blobstoreClient.Sign(objectId, action, expiration)
					Expect(err).To(HaveOccurred())
				})
			})
		})

		Context("when SwiftAuthAccount is NOT empty", func() {
			BeforeEach(func() {
				s3Config = &config.S3Cli{
					AccessKeyID:      "id",
					SecretAccessKey:  "key",
					BucketName:       "some-bucket",
					Host:             "host-name",
					SwiftAuthAccount: "swift_account",
					SwiftTempURLKey:  "temp_key",
				}

				s3Client, err := client.NewAwsS3Client(s3Config)
				Expect(err).NotTo(HaveOccurred())

				blobstoreClient = client.New(s3Client, s3Config)

				urlRegexp =
					"https://host-name/v1/swift_account/some-bucket/test-object-id" +
						`\?temp_url_sig=([a-f0-9]+)` +
						`&temp_url_expires=([0-9]+)`
			})

			Context("when the action is GET", func() {
				BeforeEach(func() {
					action = "GET"
				})

				It("returns a signed URL", func() {
					url, err := blobstoreClient.Sign(objectId, action, expiration)
					Expect(err).NotTo(HaveOccurred())

					Expect(url).To(MatchRegexp(urlRegexp))
				})
			})

			Context("when the action is PUT", func() {
				BeforeEach(func() {
					action = "PUT"
				})

				It("returns a signed URL", func() {
					url, err := blobstoreClient.Sign(objectId, action, expiration)
					Expect(err).NotTo(HaveOccurred())

					Expect(url).To(MatchRegexp(urlRegexp))
				})
			})

			Context("when the action is neither GET nor PUT", func() {
				BeforeEach(func() {
					action = "UNSUPPORTED_ACTION"
				})

				It("returns an error", func() {
					_, err := blobstoreClient.Sign(objectId, action, expiration)
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})
})
