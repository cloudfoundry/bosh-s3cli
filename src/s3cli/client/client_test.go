package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	amzaws "launchpad.net/goamz/aws"
	amzs3 "launchpad.net/goamz/s3"

	. "s3cli/client"
)

var _ = Describe("New", func() {
	Context("with basic configuration", func() {
		It("returns properly configured client", func() {
			config := Config{
				AccessKeyID:     "fake-access-key",
				SecretAccessKey: "fake-secret-key",
				BucketName:      "fake-bucket-name",
			}
			client, err := New(config)
			Expect(err).ToNot(HaveOccurred())

			expectedS3 := amzs3.New(
				amzaws.Auth{
					AccessKey: "fake-access-key",
					SecretKey: "fake-secret-key",
				},
				config.AWSRegion(),
			)

			bucket := client.(*amzs3.Bucket)
			Expect(bucket.S3).To(Equal(expectedS3))
			Expect(bucket.Name).To(Equal("fake-bucket-name"))
		})
	})

	Context("with more detailed configuration", func() {
		It("returns properly configured client", func() {
			config := Config{
				AccessKeyID:     "fake-access-key",
				SecretAccessKey: "fake-secret-key",
				BucketName:      "fake-bucket-name",
				Port:            123,
			}
			client, err := New(config)
			Expect(err).ToNot(HaveOccurred())

			expectedS3 := amzs3.New(
				amzaws.Auth{
					AccessKey: "fake-access-key",
					SecretKey: "fake-secret-key",
				},
				config.AWSRegion(),
			)

			bucket := client.(*amzs3.Bucket)
			Expect(bucket.S3).To(Equal(expectedS3))
			Expect(bucket.Name).To(Equal("fake-bucket-name"))
		})
	})
})
