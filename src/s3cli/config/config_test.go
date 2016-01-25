package config_test

import (
	"bytes"
	"errors"
	"s3cli/config"

	"encoding/json"

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
				It("reports that region and endpoint should be used for SDK configuration", func() {
					dummyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key", "bucket_name": "some-bucket", "host": "some-host", "region": "some-region"}`)
					dummyJSONReader := bytes.NewReader(dummyJSONBytes)

					c, err := config.NewFromReader(dummyJSONReader)
					Expect(err).ToNot(HaveOccurred())
					Expect(c.UseRegion()).To(BeTrue())
					Expect(c.Host).To(Equal("some-host"))
					Expect(c.Region).To(Equal("some-region"))
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

		Describe("configuring signing method", func() {
			var s3CliConfig config.S3Cli
			var configToTest map[string]interface{}
			JustBeforeEach(func() {
				configJson, err := json.Marshal(configToTest)
				Expect(err).ToNot(HaveOccurred())
				dummyJSONReader := bytes.NewReader(configJson)
				s3CliConfig, err = config.NewFromReader(dummyJSONReader)
				Expect(err).ToNot(HaveOccurred())
			})

			Context("when there is a host and no region defined", func() {
				Context("when the hostname contains 'amazonaws'", func() {
					Context("when the credential source is static", func() {
						BeforeEach(func() {
							configToTest = map[string]interface{}{
								"credentials_source": "static",
								"access_key_id":      "id",
								"secret_access_key":  "key",
								"bucket_name":        "some-bucket",
								"host":               "something.amazonaws.com.something",
							}
						})

						It("uses v2 signing", func() {
							Expect(s3CliConfig.UseV2SigningMethod).To(BeTrue())
						})
					})

					Context("when the credential source is env_or_profile", func() {
						BeforeEach(func() {
							configToTest = map[string]interface{}{
								"credentials_source": "env_or_profile",
								"bucket_name":        "some-bucket",
								"host":               "something.amazonaws.com.something",
							}
						})

						It("uses v4 signing", func() {
							Expect(s3CliConfig.UseV2SigningMethod).To(BeFalse())
						})

						Context("when signing_version is overriden", func() {
							BeforeEach(func() {
								configToTest["signature_version"] = "2"
							})

							It("uses v2 signing", func() {
								Expect(s3CliConfig.UseV2SigningMethod).To(BeTrue())
							})
						})
					})
				})

				Context("when the hostname does not contain 'amazonaws'", func() {
					BeforeEach(func() {
						configToTest = map[string]interface{}{
							"credentials_source": "static",
							"access_key_id":      "id",
							"secret_access_key":  "key",
							"bucket_name":        "some-bucket",
							"host":               "s3-compatible.com",
						}
					})

					It("uses v2 signing", func() {
						Expect(s3CliConfig.UseV2SigningMethod).To(BeTrue())
					})

					Context("when signing_version is overriden", func() {
						BeforeEach(func() {
							configToTest["signature_version"] = "4"
						})

						It("uses v4 signing", func() {
							Expect(s3CliConfig.UseV2SigningMethod).To(BeFalse())
						})
					})
				})
			})

			Context("when there is no host and no region defined", func() {
				BeforeEach(func() {
					configToTest = map[string]interface{}{
						"credentials_source": "static",
						"access_key_id":      "id",
						"secret_access_key":  "key",
						"bucket_name":        "some-bucket",
					}
				})

				It("uses v4 signing", func() {
					Expect(s3CliConfig.UseV2SigningMethod).To(BeFalse())
				})
			})

			Context("when there is no host and region defined", func() {
				BeforeEach(func() {
					configToTest = map[string]interface{}{
						"credentials_source": "static",
						"access_key_id":      "id",
						"secret_access_key":  "key",
						"bucket_name":        "some-bucket",
						"region":             "some-region-3",
					}
				})

				It("uses v4 signing", func() {
					Expect(s3CliConfig.UseV2SigningMethod).To(BeFalse())
				})
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
			It("returns a empty string URI if `host` is empty", func() {
				dummyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key", "bucket_name": "some-bucket", "host": "", "port": 443}`)
				dummyJSONReader := bytes.NewReader(dummyJSONBytes)

				c, err := config.NewFromReader(dummyJSONReader)
				Expect(err).ToNot(HaveOccurred())
				Expect(c.S3Endpoint()).To(Equal(""))
			})
		})

		Context("when port is not provided", func() {
			It("returns a URI in the form `host` only", func() {
				dummyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key", "bucket_name": "some-bucket", "host": "some-host-name"}`)
				dummyJSONReader := bytes.NewReader(dummyJSONBytes)

				c, err := config.NewFromReader(dummyJSONReader)
				Expect(err).ToNot(HaveOccurred())
				Expect(c.S3Endpoint()).To(Equal("some-host-name"))
			})
			It("returns a empty string URI if `host` is empty", func() {
				dummyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key", "bucket_name": "some-bucket", "host": ""}`)
				dummyJSONReader := bytes.NewReader(dummyJSONBytes)

				c, err := config.NewFromReader(dummyJSONReader)
				Expect(err).ToNot(HaveOccurred())
				Expect(c.S3Endpoint()).To(Equal(""))
			})
		})
	})

	Describe("validating credentials", func() {
		Describe("when credentials source is not specified", func() {
			Context("when a secret key and access key are provided", func() {
				It("defaults to static credentials", func() {
					dummyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key", "bucket_name": "some-bucket"}`)
					dummyJSONReader := bytes.NewReader(dummyJSONBytes)

					c, err := config.NewFromReader(dummyJSONReader)
					Expect(err).ToNot(HaveOccurred())
					Expect(c.CredentialsSource).To(Equal("static"))
				})
			})

			Context("when either the secret key or access key are missing", func() {
				It("raises an error", func() {
					dummyJSONBytes := []byte(`{"secret_access_key": "key", "bucket_name": "some-bucket"}`)
					dummyJSONReader := bytes.NewReader(dummyJSONBytes)

					_, err := config.NewFromReader(dummyJSONReader)
					Expect(err).To(MatchError("access_key_id and secret_access_key must be provided"))
				})
			})

			Context("when neither an access key or secret key are provided", func() {
				It("defaults credentials source to anonymous", func() {
					dummyJSONBytes := []byte(`{"bucket_name": "some-bucket"}`)
					dummyJSONReader := bytes.NewReader(dummyJSONBytes)

					c, err := config.NewFromReader(dummyJSONReader)
					Expect(err).ToNot(HaveOccurred())
					Expect(c.CredentialsSource).To(Equal("none"))
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

		})

		Context("when credential source is `static`", func() {
			It("validates that access key and secret key are set", func() {
				dummyJSONBytes := []byte(`{"bucket_name": "some-bucket", "access_key_id": "some_id"}`)
				dummyJSONReader := bytes.NewReader(dummyJSONBytes)
				_, err := config.NewFromReader(dummyJSONReader)
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

		Context("when the credentials source is `none`", func() {
			It("validates that access key and secret key are not set", func() {
				dummyJSONBytes := []byte(`{"bucket_name": "some-bucket", "credentials_source": "none", "access_key_id": "some_id"}`)
				dummyJSONReader := bytes.NewReader(dummyJSONBytes)
				_, err := config.NewFromReader(dummyJSONReader)
				Expect(err).To(MatchError("can't use access_key_id and secret_access_key with none credentials_source"))
			})
		})
	})
})

type explodingReader struct{}

func (e explodingReader) Read([]byte) (int, error) {
	return 0, errors.New("explosion")
}
