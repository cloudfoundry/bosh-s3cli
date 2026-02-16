package config_test

import (
	"bytes"
	"errors"

	"github.com/cloudfoundry/bosh-s3cli/config"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("BlobstoreClient configuration", func() {
	DescribeTable("Provider",
		func(host, provider string) {
			Expect(config.Provider(host)).To(Equal(provider))
		},
		Entry("aws 1", "s3.amazonaws.com", "aws"),
		Entry("aws 2", "s3.external-1.amazonaws.com", "aws"),
		Entry("aws 3", "s3.some-region.amazonaws.com", "aws"),
		Entry("alicloud 1", "oss-r-s-1-internal.aliyuncs.com", "alicloud"),
		Entry("alicloud 2", "oss-r-s-internal.aliyuncs.com", "alicloud"),
		Entry("alicloud 3", "oss-r-s-1.aliyuncs.com", "alicloud"),
		Entry("alicloud 4", "oss-r-s.aliyuncs.com", "alicloud"),
		Entry("google 1", "storage.googleapis.com", "google"),
	)

	Describe("building a configuration", func() {
		Describe("checking that either host or region has been set", func() {

			Context("when AWS endpoint has been set but not region", func() {

				It("sets the AWS region based on the hostname", func() {
					dummyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key", "bucket_name": "some-bucket", "host": "s3.amazonaws.com"}`)
					dummyJSONReader := bytes.NewReader(dummyJSONBytes)
					c, err := config.NewFromReader(dummyJSONReader)
					Expect(err).ToNot(HaveOccurred())
					Expect(c.Host).To(Equal("s3.amazonaws.com"))
					Expect(c.Region).To(Equal("us-east-1"))
				})
			})

			Context("when Google endpoint has been set but not region", func() {
				dummyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key", "bucket_name": "some-bucket", "host": "storage.googleapis.com"}`)
				dummyJSONReader := bytes.NewReader(dummyJSONBytes)

				It("sets the default region used for SDK configuration", func() {
					c, err := config.NewFromReader(dummyJSONReader)
					Expect(err).ToNot(HaveOccurred())
					Expect(c.Host).To(Equal("storage.googleapis.com"))
					Expect(c.Region).To(Equal("us-east-1"))
				})
			})

			Context("when Ali endpoint with region has been set but without explicit region", func() {
				dummyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key", "bucket_name": "some-bucket", "host": "oss-some-region-internal.aliyuncs.com"}`)
				dummyJSONReader := bytes.NewReader(dummyJSONBytes)

				It("parses region from host for SDK configuration", func() {
					c, err := config.NewFromReader(dummyJSONReader)
					Expect(err).ToNot(HaveOccurred())
					Expect(c.Host).To(Equal("oss-some-region-internal.aliyuncs.com"))
					Expect(c.Region).To(Equal("some-region"))
				})
			})

			Context("when region has been set but not host", func() {
				dummyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key", "bucket_name": "some-bucket", "region": "some-region"}`)
				dummyJSONReader := bytes.NewReader(dummyJSONBytes)

				It("reports that region should be used for SDK configuration", func() {
					c, err := config.NewFromReader(dummyJSONReader)
					Expect(err).ToNot(HaveOccurred())
					Expect(c.Host).To(Equal(""))
					Expect(c.Region).To(Equal("some-region"))
				})
			})

			Context("when non-AWS host and region have been set", func() {
				dummyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key", "bucket_name": "some-bucket", "host": "some-host", "region": "some-region"}`)
				dummyJSONReader := bytes.NewReader(dummyJSONBytes)

				It("sets region and endpoint to user-specified values", func() {
					c, err := config.NewFromReader(dummyJSONReader)
					Expect(err).ToNot(HaveOccurred())
					Expect(c.Host).To(Equal("some-host"))
					Expect(c.Region).To(Equal("some-region"))
				})
			})

			Context("when AWS host and region have been set", func() {
				dummyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key", "bucket_name": "some-bucket", "host": "s3.amazonaws.com", "region": "us-west-1"}`)
				dummyJSONReader := bytes.NewReader(dummyJSONBytes)

				It("does not override the user-specified region based on the hostname", func() {
					c, err := config.NewFromReader(dummyJSONReader)
					Expect(err).ToNot(HaveOccurred())
					Expect(c.Host).To(Equal("s3.amazonaws.com"))
					Expect(c.Region).To(Equal("us-west-1"))
				})
			})

			Context("when neither host and region have been set", func() {
				dummyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key", "bucket_name": "some-bucket"}`)
				dummyJSONReader := bytes.NewReader(dummyJSONBytes)

				It("defaults region to us-east-1", func() {
					c, err := config.NewFromReader(dummyJSONReader)
					Expect(err).ToNot(HaveOccurred())
					Expect(c.Host).To(Equal(""))
					Expect(c.Region).To(Equal("us-east-1"))
				})
			})

			Context("when MultipartUpload have been set", func() {
				dummyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key", "bucket_name": "some-bucket", "host": "some-host", "region": "some-region", "multipart_upload": false}`)
				dummyJSONReader := bytes.NewReader(dummyJSONBytes)
				It("sets MultipartUpload to user-specified values", func() {
					c, err := config.NewFromReader(dummyJSONReader)
					Expect(err).ToNot(HaveOccurred())
					Expect(c.MultipartUpload).To(BeFalse())
				})
			})

			Context("when MultipartUpload have not been set", func() {
				dummyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key", "bucket_name": "some-bucket", "host": "some-host", "region": "some-region"}`)
				dummyJSONReader := bytes.NewReader(dummyJSONBytes)
				It("default MultipartUpload to true", func() {
					c, err := config.NewFromReader(dummyJSONReader)
					Expect(err).ToNot(HaveOccurred())
					Expect(c.MultipartUpload).To(BeTrue())
				})
			})

			Context("when HostStyle has been set", func() {
				dummyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key", "bucket_name": "some-bucket", "host": "some-host", "region": "some-region", "host_style": true}`)
				dummyJSONReader := bytes.NewReader(dummyJSONBytes)
				It("sets HostStyle to user-specified value", func() {
					c, err := config.NewFromReader(dummyJSONReader)
					Expect(err).ToNot(HaveOccurred())
					Expect(c.HostStyle).To(BeTrue())
				})
			})
		})

		Describe("when bucket is not specified", func() {
			emptyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key"}`)
			emptyJSONReader := bytes.NewReader(emptyJSONBytes)

			It("returns an error", func() {
				_, err := config.NewFromReader(emptyJSONReader)
				Expect(err).To(MatchError("bucket_name must be set"))
			})
		})

		Describe("when bucket is specified", func() {
			emptyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key", "bucket_name": "some-bucket"}`)
			emptyJSONReader := bytes.NewReader(emptyJSONBytes)

			It("uses the given bucket", func() {
				c, err := config.NewFromReader(emptyJSONReader)
				Expect(err).ToNot(HaveOccurred())
				Expect(c.BucketName).To(Equal("some-bucket"))
			})
		})

		Describe("when folder is specified", func() {
			emptyJSONBytes := []byte(`{
				"access_key_id": "id",
				"secret_access_key": "key",
				"bucket_name": "some-bucket",
				"folder_name": "some-folder/other-folder"
			}`)
			emptyJSONReader := bytes.NewReader(emptyJSONBytes)

			It("uses the given folder", func() {
				c, err := config.NewFromReader(emptyJSONReader)
				Expect(err).ToNot(HaveOccurred())
				Expect(c.FolderName).To(Equal("some-folder/other-folder"))
			})
		})

		Describe("Default SSL options", func() {
			emptyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key", "bucket_name": "some-bucket"}`)
			emptyJSONReader := bytes.NewReader(emptyJSONBytes)

			It("defaults to use SSL and peer verification", func() {
				c, err := config.NewFromReader(emptyJSONReader)
				Expect(err).ToNot(HaveOccurred())
				Expect(c.UseSSL).To(BeTrue())
				Expect(c.SSLVerifyPeer).To(BeTrue())
			})
		})

		Describe("configing force path style", func() {
			It("when Alibaba Cloud provider", func() {
				configBytes := []byte(`{
					"access_key_id":      "id",
					"secret_access_key":  "key",
					"bucket_name":        "some-bucket",
					"host":               "oss-some-region.aliyuncs.com"
				}`)

				configReader := bytes.NewReader(configBytes)
				s3CliConfig, err := config.NewFromReader(configReader)
				Expect(err).ToNot(HaveOccurred())
				Expect(s3CliConfig.HostStyle).To(BeTrue())
			})

			It("when AWS provider", func() {
				configBytes := []byte(`{
					"access_key_id":      "id",
					"secret_access_key":  "key",
					"bucket_name":        "some-bucket",
					"host": 	      "s3.amazonaws.com"
				}`)

				configReader := bytes.NewReader(configBytes)
				s3CliConfig, err := config.NewFromReader(configReader)
				Expect(err).ToNot(HaveOccurred())
				Expect(s3CliConfig.HostStyle).To(BeFalse())
			})

			It("when Google provider", func() {
				configBytes := []byte(`{
					"access_key_id":      "id",
					"secret_access_key":  "key",
					"bucket_name":        "some-bucket",
					"host": 	      "storage.googleapis.com"
				}`)

				configReader := bytes.NewReader(configBytes)
				s3CliConfig, err := config.NewFromReader(configReader)
				Expect(err).ToNot(HaveOccurred())
				Expect(s3CliConfig.HostStyle).To(BeFalse())
			})

			It("when Default provider", func() {
				configBytes := []byte(`{
					"access_key_id":      "id",
					"secret_access_key":  "key",
					"bucket_name":        "some-bucket",
					"host": 	      "storage.googleapis.com"
				}`)

				configReader := bytes.NewReader(configBytes)
				s3CliConfig, err := config.NewFromReader(configReader)
				Expect(err).ToNot(HaveOccurred())
				Expect(s3CliConfig.HostStyle).To(BeFalse())
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

	Describe("transfer tuning fields (concurrency / part size)", func() {
		It("accepts zero as 'use defaults' and preserves zeros when not set", func() {
			dummyJSONBytes := []byte(`{"access_key_id":"id","secret_access_key":"key","bucket_name":"some-bucket"}`)
			dummyJSONReader := bytes.NewReader(dummyJSONBytes)

			c, err := config.NewFromReader(dummyJSONReader)
			Expect(err).ToNot(HaveOccurred())
			Expect(c.DownloadConcurrency).To(Equal(0))
			Expect(c.UploadConcurrency).To(Equal(0))
			Expect(c.DownloadPartSize).To(Equal(int64(0)))
			Expect(c.UploadPartSize).To(Equal(int64(0)))
		})

		It("preserves positive tuning values from config", func() {
			dummyJSONBytes := []byte(`{
				"access_key_id":"id",
				"secret_access_key":"key",
				"bucket_name":"some-bucket",
				"download_concurrency": 10,
				"download_part_size": 10485760,
				"upload_concurrency": 8,
				"upload_part_size": 5242880
			}`)
			dummyJSONReader := bytes.NewReader(dummyJSONBytes)

			c, err := config.NewFromReader(dummyJSONReader)
			Expect(err).ToNot(HaveOccurred())
			Expect(c.DownloadConcurrency).To(Equal(10))
			Expect(c.DownloadPartSize).To(Equal(int64(10485760)))
			Expect(c.UploadConcurrency).To(Equal(8))
			Expect(c.UploadPartSize).To(Equal(int64(5242880)))
		})

		It("rejects negative tuning values", func() {
			dummyJSONBytes := []byte(`{
				"access_key_id":"id",
				"secret_access_key":"key",
				"bucket_name":"some-bucket",
				"download_concurrency": -1
			}`)
			dummyJSONReader := bytes.NewReader(dummyJSONBytes)

			_, err := config.NewFromReader(dummyJSONReader)
			Expect(err).To(MatchError("download/upload concurrency and part sizes must be non-negative"))

			// negative part size
			dummyJSONBytes = []byte(`{
				"access_key_id":"id",
				"secret_access_key":"key",
				"bucket_name":"some-bucket",
				"upload_part_size": -1024
			}`)
			dummyJSONReader = bytes.NewReader(dummyJSONBytes)

			_, err = config.NewFromReader(dummyJSONReader)
			Expect(err).To(MatchError("download/upload concurrency and part sizes must be non-negative"))
		})
	})

	Describe("returning the S3 endpoint", func() {
		Context("when port is provided", func() {
			It("returns a URI in the form `host:port`", func() {
				dummyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key", "bucket_name": "some-bucket", "use_ssl": false, "host": "some-host-name", "port": 443}`)
				dummyJSONReader := bytes.NewReader(dummyJSONBytes)

				c, err := config.NewFromReader(dummyJSONReader)
				Expect(err).ToNot(HaveOccurred())
				Expect(c.S3Endpoint()).To(Equal("some-host-name:443"))
			})
			It("returns a URI in the form `host` when protocol and port match", func() {
				dummyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key", "bucket_name": "some-bucket", "use_ssl": true, "host": "some-host-name", "port": 443}`)
				dummyJSONReader := bytes.NewReader(dummyJSONBytes)

				c, err := config.NewFromReader(dummyJSONReader)
				Expect(err).ToNot(HaveOccurred())
				Expect(c.S3Endpoint()).To(Equal("some-host-name"))

				dummyJSONBytes = []byte(`{"access_key_id": "id", "secret_access_key": "key", "bucket_name": "some-bucket", "use_ssl": false, "host": "some-host-name", "port": 80}`)
				dummyJSONReader = bytes.NewReader(dummyJSONBytes)

				c, err = config.NewFromReader(dummyJSONReader)
				Expect(err).ToNot(HaveOccurred())
				Expect(c.S3Endpoint()).To(Equal("some-host-name"))
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
					Expect(err).To(MatchError("invalid credentials_source: magical_unicorns"))
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

	Describe("returning the alibaba cloud region", func() {
		Context("when host is provided", func() {
			It("returns a region id in the public `host`", func() {
				dummyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key", "bucket_name": "some-bucket", "host": "oss-some-region.aliyuncs.com"}`)
				dummyJSONReader := bytes.NewReader(dummyJSONBytes)

				c, err := config.NewFromReader(dummyJSONReader)
				Expect(err).ToNot(HaveOccurred())
				Expect(c.Region).To(Equal("some-region"))
			})
			It("returns a region id in the private `host`", func() {
				dummyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key", "bucket_name": "some-bucket", "host": "oss-some-region-internal.aliyuncs.com"}`)
				dummyJSONReader := bytes.NewReader(dummyJSONBytes)

				c, err := config.NewFromReader(dummyJSONReader)
				Expect(err).ToNot(HaveOccurred())
				Expect(c.Region).To(Equal("some-region"))
			})
			It("returns a empty string if `host` is empty", func() {
				dummyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key", "bucket_name": "some-bucket", "host": ""}`)
				dummyJSONReader := bytes.NewReader(dummyJSONBytes)

				c, err := config.NewFromReader(dummyJSONReader)
				Expect(err).ToNot(HaveOccurred())
				Expect(c.S3Endpoint()).To(Equal(""))
			})
		})
	})

	Describe("returning the alibaba cloud endpoint", func() {
		Context("when port is provided", func() {
			It("returns a URI in the form `host:port`", func() {
				dummyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key", "bucket_name": "some-bucket", "use_ssl": false, "host": "oss-some-region.aliyuncs.com", "port": 443}`)
				dummyJSONReader := bytes.NewReader(dummyJSONBytes)

				c, err := config.NewFromReader(dummyJSONReader)
				Expect(err).ToNot(HaveOccurred())
				Expect(c.S3Endpoint()).To(Equal("oss-some-region.aliyuncs.com:443"))
				Expect(c.Host).To(Equal("oss-some-region.aliyuncs.com"))
			})
			It("returns a empty string URI if `host` is empty", func() {
				dummyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key", "bucket_name": "some-bucket", "host": "", "port": 443}`)
				dummyJSONReader := bytes.NewReader(dummyJSONBytes)

				c, err := config.NewFromReader(dummyJSONReader)
				Expect(err).ToNot(HaveOccurred())
				Expect(c.S3Endpoint()).To(Equal(""))
				Expect(c.Host).To(Equal(""))
			})
		})

		Context("when port is not provided", func() {
			It("returns a URI in the form `host` only", func() {
				dummyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key", "bucket_name": "some-bucket", "host": "oss-some-region.aliyuncs.com"}`)
				dummyJSONReader := bytes.NewReader(dummyJSONBytes)

				c, err := config.NewFromReader(dummyJSONReader)
				Expect(err).ToNot(HaveOccurred())
				Expect(c.S3Endpoint()).To(Equal("oss-some-region.aliyuncs.com"))
				Expect(c.Host).To(Equal("oss-some-region.aliyuncs.com"))
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

	Describe("checking the alibaba cloud MultipartUpload", func() {
		emptyJSONBytes := []byte(`{"access_key_id": "id", "secret_access_key": "key", "bucket_name": "some-bucket", "host": "oss-some-region.aliyuncs.com"}`)
		emptyJSONReader := bytes.NewReader(emptyJSONBytes)

		It("defaults to support multipart uploading", func() {
			c, err := config.NewFromReader(emptyJSONReader)
			Expect(err).ToNot(HaveOccurred())
			Expect(c.MultipartUpload).To(BeTrue())
		})
	})
})

type explodingReader struct{}

func (e explodingReader) Read([]byte) (int, error) {
	return 0, errors.New("explosion")
}
