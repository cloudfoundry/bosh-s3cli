package config_test

import (
	"bytes"
	"errors"
	"s3cli/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BlobstoreClient configuration", func() {
	Describe("ignoring region configuration", func() {
		It("allows for the S3 SDK to be configured with empty region information", func() {
			Expect(config.EmptyRegion).To(Equal(" "))
		})
	})

	Describe("building a configuration", func() {
		Describe("checking that either host or region has been set", func() {
			Context("when host has been set but not region", func() {
				It("reports that region should not be used for SDK configuration", func() {
					dummyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key", "bucket_name": "some-bucket", "host": "some-host"}`)
					dummyJSONReader := bytes.NewReader(dummyJSONBytes)

					c, err := config.NewFromReader(dummyJSONReader)
					Expect(err).ToNot(HaveOccurred())
					Expect(c.UseRegion()).To(BeFalse())
					Expect(c.Host).To(Equal("some-host"))
					Expect(c.Region).To(Equal(""))
				})
			})

			Context("when region has been set but not host", func() {
				It("reports that region should be used for SDK configuration", func() {
					dummyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key", "bucket_name": "some-bucket", "region": "some-region"}`)
					dummyJSONReader := bytes.NewReader(dummyJSONBytes)

					c, err := config.NewFromReader(dummyJSONReader)
					Expect(err).ToNot(HaveOccurred())
					Expect(c.UseRegion()).To(BeTrue())
					Expect(c.Host).To(Equal(""))
					Expect(c.Region).To(Equal("some-region"))
				})
			})

			Context("when both host and region have been set", func() {
				It("returns an error", func() {
					dummyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key", "bucket_name": "some-bucket", "host": "some-host", "region": "some-region"}`)
					dummyJSONReader := bytes.NewReader(dummyJSONBytes)

					_, err := config.NewFromReader(dummyJSONReader)
					Expect(err).To(MatchError("Cannot set both region and host at the same time"))
				})
			})

			Context("when neither host and region have been set", func() {
				It("defaults region to us-east-1", func() {
					dummyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key", "bucket_name": "some-bucket"}`)
					dummyJSONReader := bytes.NewReader(dummyJSONBytes)

					c, err := config.NewFromReader(dummyJSONReader)
					Expect(err).ToNot(HaveOccurred())
					Expect(c.Host).To(Equal(""))
					Expect(c.Region).To(Equal("us-east-1"))
				})
			})
		})

		Describe("when bucket is not specified", func() {
			It("returns an error", func() {
				emptyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key"}`)
				emptyJSONReader := bytes.NewReader(emptyJSONBytes)

				_, err := config.NewFromReader(emptyJSONReader)
				Expect(err).To(MatchError("bucket_name must be set"))
			})
		})

		Describe("when bucket is specified", func() {
			It("uses the given bucket", func() {
				emptyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key", "bucket_name": "some-bucket"}`)
				emptyJSONReader := bytes.NewReader(emptyJSONBytes)

				c, err := config.NewFromReader(emptyJSONReader)
				Expect(err).ToNot(HaveOccurred())
				Expect(c.BucketName).To(Equal("some-bucket"))
			})
		})

		Describe("Default SSL options", func() {
			It("defaults to use SSL and peer verification", func() {
				emptyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key", "bucket_name": "some-bucket"}`)
				emptyJSONReader := bytes.NewReader(emptyJSONBytes)

				c, err := config.NewFromReader(emptyJSONReader)
				Expect(err).ToNot(HaveOccurred())
				Expect(c.UseSSL).To(BeTrue())
				Expect(c.SSLVerifyPeer).To(BeTrue())
			})
		})

		Describe("when credentials source is not specified", func() {
			It("defaults credentials source to static", func() {
				dummyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key", "bucket_name": "some-bucket"}`)
				dummyJSONReader := bytes.NewReader(dummyJSONBytes)

				c, err := config.NewFromReader(dummyJSONReader)
				Expect(err).ToNot(HaveOccurred())
				Expect(c.CredentialsSource).To(Equal("static"))
			})
		})

		Describe("when credentials source is invalid", func() {
			It("returns an error", func() {
				dummyJSONBytes := []byte(`{"bucket_name": "some-bucket", "credentials_source": "magical_unicorns"}`)
				dummyJSONReader := bytes.NewReader(dummyJSONBytes)

				_, err := config.NewFromReader(dummyJSONReader)
				Expect(err).To(MatchError("Invalid credentials_source: magical_unicorns"))
			})
		})

		Describe("configuring signing method", func() {
			It("uses v4 signing in the `eu-central-1` AWS region", func() {
				dummyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key", "bucket_name": "some-bucket", "region": "eu-central-1"}`)
				dummyJSONReader := bytes.NewReader(dummyJSONBytes)

				c, err := config.NewFromReader(dummyJSONReader)
				Expect(err).ToNot(HaveOccurred())
				Expect(c.UseV2SigningMethod).To(BeFalse())
			})

			It("uses v4 signing in the `cn-north-1` AWS region", func() {
				dummyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key", "bucket_name": "some-bucket", "region": "cn-north-1"}`)
				dummyJSONReader := bytes.NewReader(dummyJSONBytes)

				c, err := config.NewFromReader(dummyJSONReader)
				Expect(err).ToNot(HaveOccurred())
				Expect(c.UseV2SigningMethod).To(BeFalse())
			})

			It("uses v2 signing for any other regions", func() {
				regions := []string{"us-east-1", "us-west-1", "unknown-region"}

				for _, region := range regions {
					dummyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key", "bucket_name": "some-bucket", "region": "` + region + `"}`)
					dummyJSONReader := bytes.NewReader(dummyJSONBytes)

					c, err := config.NewFromReader(dummyJSONReader)
					Expect(err).ToNot(HaveOccurred())
					Expect(c.UseV2SigningMethod).To(BeTrue())
				}
			})

			It("defaults to v2 signing if region is not specified", func() {
				dummyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key", "bucket_name": "some-bucket"}`)
				dummyJSONReader := bytes.NewReader(dummyJSONBytes)

				c, err := config.NewFromReader(dummyJSONReader)
				Expect(err).ToNot(HaveOccurred())
				Expect(c.UseV2SigningMethod).To(BeTrue())
			})
		})

		Context("when the configuration file cannot be read", func() {
			It("returns an error", func() {
				f := explodingReader{}

				_, err := config.NewFromReader(f)
				Expect(err).To(MatchError("explosion"))
			})
		})

		Context("when the configuration file is invalid JSON", func() {
			It("returns an error", func() {
				invalidJSONBytes := []byte(`invalid-json`)
				invalidJSONReader := bytes.NewReader(invalidJSONBytes)

				_, err := config.NewFromReader(invalidJSONReader)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("returning the S3 endpoint", func() {
		Context("when port is provided", func() {
			It("returns a URI in the form `host:port`", func() {
				dummyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key", "bucket_name": "some-bucket", "host": "some-host-name", "port": 443}`)
				dummyJSONReader := bytes.NewReader(dummyJSONBytes)

				c, err := config.NewFromReader(dummyJSONReader)
				Expect(err).ToNot(HaveOccurred())
				Expect(c.S3Endpoint()).To(Equal("some-host-name:443"))
			})

			Context("when port is not provided", func() {
				It("returns a URI in the form `host` only", func() {
					dummyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key", "bucket_name": "some-bucket", "host": "some-host-name"}`)
					dummyJSONReader := bytes.NewReader(dummyJSONBytes)

					c, err := config.NewFromReader(dummyJSONReader)
					Expect(err).ToNot(HaveOccurred())
					Expect(c.S3Endpoint()).To(Equal("some-host-name"))
				})
			})
		})
	})

	Describe("validating credentials", func() {
		Context("when credential source is `static`", func() {
			It("validates that access key and secret key are set", func() {
				dummyJSONBytes := []byte(`{"bucket_name": "some-bucket"}`)
				dummyJSONReader := bytes.NewReader(dummyJSONBytes)

				_, err := config.NewFromReader(dummyJSONReader)
				Expect(err).To(MatchError("access_key_id and secret_access_key must be provided"))

				dummyJSONBytes = []byte(`{"bucket_name": "some-bucket", "access_key_id": "some_id"}`)
				dummyJSONReader = bytes.NewReader(dummyJSONBytes)
				_, err = config.NewFromReader(dummyJSONReader)
				Expect(err).To(MatchError("access_key_id and secret_access_key must be provided"))

				dummyJSONBytes = []byte(`{"bucket_name": "some-bucket", "access_key_id": "some_id", "secret_access_key": "some_secret"}`)
				dummyJSONReader = bytes.NewReader(dummyJSONBytes)
				_, err = config.NewFromReader(dummyJSONReader)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when credentials source is `env_or_profile`", func() {
			It("validates that access key and secret key are not set", func() {
				dummyJSONBytes := []byte(`{"bucket_name": "some-bucket", "credentials_source": "env_or_profile"}`)
				dummyJSONReader := bytes.NewReader(dummyJSONBytes)

				_, err := config.NewFromReader(dummyJSONReader)
				Expect(err).ToNot(HaveOccurred())

				dummyJSONBytes = []byte(`{"bucket_name": "some-bucket", "credentials_source": "env_or_profile", "access_key_id": "some_id"}`)
				dummyJSONReader = bytes.NewReader(dummyJSONBytes)
				_, err = config.NewFromReader(dummyJSONReader)
				Expect(err).To(MatchError("can't use access_key_id and secret_access_key with env_or_profile credentials_source"))

				dummyJSONBytes = []byte(`{"bucket_name": "some-bucket", "credentials_source": "env_or_profile", "access_key_id": "some_id", "secret_access_key": "some_secret"}`)
				dummyJSONReader = bytes.NewReader(dummyJSONBytes)
				_, err = config.NewFromReader(dummyJSONReader)
				Expect(err).To(MatchError("can't use access_key_id and secret_access_key with env_or_profile credentials_source"))
			})
		})
	})
})

type explodingReader struct{}

func (e explodingReader) Read([]byte) (int, error) {
	return 0, errors.New("explosion")
}
