package client_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	amzaws "gopkg.in/amz.v3/aws"
	amzs3 "gopkg.in/amz.v3/s3"

	. "s3cli/client"
)

type HTTPHandler func(http.ResponseWriter, *http.Request)

func (h HTTPHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) { h(rw, r) }

var _ = Describe("New", func() {
	Context("with basic configuration", func() {
		It("returns properly configured client", func() {
			config := Config{
				AccessKeyID:       "fake-access-key",
				SecretAccessKey:   "fake-secret-key",
				BucketName:        "fake-bucket-name",
				CredentialsSource: "static",
				Region:            "fake-region-name",
			}
			client, err := New(config)
			Expect(err).ToNot(HaveOccurred())

			expectedS3 := amzs3.New(
				amzaws.Auth{
					AccessKey: "fake-access-key",
					SecretKey: "fake-secret-key",
				},
				config.AWSRegion(),
				nil,
			)

			bucket := client.(*amzs3.Bucket)
			Expect(bucket.S3.Auth).To(Equal(expectedS3.Auth))
			Expect(bucket.S3.Region).To(Equal(expectedS3.Region))
			Expect(reflect.TypeOf(bucket.S3.Sign)).To(Equal(reflect.TypeOf(expectedS3.Sign)))
			Expect(bucket.Name).To(Equal("fake-bucket-name"))
		})
	})

	Context("with more detailed configuration", func() {
		It("returns properly configured client", func() {
			config := Config{
				AccessKeyID:       "fake-access-key",
				SecretAccessKey:   "fake-secret-key",
				BucketName:        "fake-bucket-name",
				CredentialsSource: "static",
				Region:            "fake-region-name",
				Port:              123,
			}
			client, err := New(config)
			Expect(err).ToNot(HaveOccurred())

			expectedS3 := amzs3.New(
				amzaws.Auth{
					AccessKey: "fake-access-key",
					SecretKey: "fake-secret-key",
				},
				config.AWSRegion(),
				nil,
			)

			bucket := client.(*amzs3.Bucket)
			Expect(bucket.S3.Auth).To(Equal(expectedS3.Auth))
			Expect(bucket.S3.Region).To(Equal(expectedS3.Region))
			Expect(reflect.TypeOf(bucket.S3.Sign)).To(Equal(reflect.TypeOf(expectedS3.Sign)))
			Expect(bucket.Name).To(Equal("fake-bucket-name"))
		})
	})

	Context("with SSL cert verification turned off", func() {
		var server *httptest.Server

		BeforeEach(func() {
			server = httptest.NewTLSServer(
				HTTPHandler(func(w http.ResponseWriter, req *http.Request) {
					req.Body.Close()
					w.WriteHeader(200)
				}),
			)
		})

		AfterEach(func() {
			server.Close()
		})

		It("returns client that does not error when SSL cert cannot be verified", func() {
			url, err := url.Parse(server.URL)
			Expect(err).ToNot(HaveOccurred())

			parts := strings.SplitN(url.Host, ":", 2)
			Expect(len(parts)).To(Equal(2))

			hostName, portStr := parts[0], parts[1]
			port, err := strconv.Atoi(portStr)
			Expect(err).ToNot(HaveOccurred())

			config := Config{
				AccessKeyID:       "fake-access-key",
				SecretAccessKey:   "fake-secret-key",
				BucketName:        "fake-bucket-name",
				CredentialsSource: "static",
				Host:              hostName,
				Port:              port,
				UseSSL:            true,
				SSLVerifyPeer:     true,
			}

			client, err := New(config)
			Expect(err).ToNot(HaveOccurred())

			_, err = client.GetReader("fake-path")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("certificate signed by unknown authority"))

			// Make sure that requests do work after
			// SSLVerifyPeer is turned on against the *same* server
			config.SSLVerifyPeer = false

			client, err = New(config)
			Expect(err).ToNot(HaveOccurred())

			_, err = client.GetReader("fake-path")
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("with credentials_source set to static", func() {
		It("raises an error when access_key_id or secret_access_key is not provided", func() {
			config := Config{
				CredentialsSource: "static",
			}

			_, err := New(config)
			Expect(err).To(HaveOccurred())
		})

		It("returns properly configured client when access_key_id and secret_access_key are provided", func() {
			config := Config{
				AccessKeyID:       "fake-access-key",
				SecretAccessKey:   "fake-secret-key",
				BucketName:        "fake-bucket-name",
				CredentialsSource: "static",
			}

			client, err := New(config)
			Expect(err).ToNot(HaveOccurred())

			expectedS3 := amzs3.New(
				amzaws.Auth{
					AccessKey: "fake-access-key",
					SecretKey: "fake-secret-key",
				},
				config.AWSRegion(),
				nil,
			)

			bucket := client.(*amzs3.Bucket)
			Expect(bucket.S3.Auth).To(Equal(expectedS3.Auth))
			Expect(bucket.S3.Region).To(Equal(expectedS3.Region))
			Expect(bucket.S3.Sign).To(BeAssignableToTypeOf(amzaws.SignS3))
			Expect(bucket.Name).To(Equal("fake-bucket-name"))
		})
	})

	Context("with credentials_source set to env_or_profile", func() {
		It("returns properly configured client when environment variable being set", func() {
			config := Config{
				BucketName:        "fake-bucket-name",
				CredentialsSource: "env_or_profile",
			}
			os.Setenv("AWS_ACCESS_KEY_ID", "fake-access-key")
			os.Setenv("AWS_SECRET_ACCESS_KEY", "fake-secret-key")

			client, err := New(config)
			Expect(err).ToNot(HaveOccurred())

			expectedS3 := amzs3.New(
				amzaws.Auth{
					AccessKey: "fake-access-key",
					SecretKey: "fake-secret-key",
				},
				config.AWSRegion(),
				nil,
			)

			bucket := client.(*amzs3.Bucket)
			Expect(bucket.S3.Auth).To(Equal(expectedS3.Auth))
			Expect(bucket.S3.Region).To(Equal(expectedS3.Region))
			Expect(bucket.S3.Sign).To(BeAssignableToTypeOf(amzaws.SignS3))
			Expect(bucket.Name).To(Equal("fake-bucket-name"))
		})

		It("returns properly configured client when retrieving credentials from metadata", func() {
			config := Config{
				BucketName:        "fake-bucket-name",
				CredentialsSource: "env_or_profile",
			}

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.String() == "/iam/security-credentials/" {
					fmt.Fprintln(w, "fake-iam-role")
				} else {
					fmt.Fprintln(w, `{ "Code" : "Success", "LastUpdated" : "2015-09-06T10:12:18Z", "Type" : "AWS-HMAC", 
						"AccessKeyId" : "fake-access-key", "SecretAccessKey" : "fake-secret-key", 
						"Expiration" : "2015-09-06T16:47:38Z" }`)
				}
			}))

			defer ts.Close()

			amzaws.DEFAULT_METADATA_URL = ts.URL + "/"
			client, err := New(config)
			Expect(err).ToNot(HaveOccurred())

			expectedS3 := amzs3.New(
				amzaws.Auth{
					AccessKey: "fake-access-key",
					SecretKey: "fake-secret-key",
				},
				config.AWSRegion(),
				nil,
			)

			bucket := client.(*amzs3.Bucket)
			Expect(bucket.S3.Auth).To(Equal(expectedS3.Auth))
			Expect(bucket.S3.Region).To(Equal(expectedS3.Region))
			Expect(bucket.S3.Sign).To(BeAssignableToTypeOf(amzaws.SignS3))
			Expect(bucket.Name).To(Equal("fake-bucket-name"))
		})

		It("raises an error when access_key_id are also provided", func() {
			config := Config{
				AccessKeyID:       "fake-access-key",
				SecretAccessKey:   "fake-secret-key",
				BucketName:        "fake-bucket-name",
				CredentialsSource: "env_or_profile",
			}

			_, err := New(config)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("with credentials_source set to incorrect value", func() {
		It("raises an error when access_key_id are also provided", func() {
			config := Config{
				CredentialsSource: "incorrect-value",
			}

			_, err := New(config)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("with different signature version", func() {
		It("returns default s3 legacy signature", func() {
			config := Config{
				AccessKeyID:       "fake-access-key",
				SecretAccessKey:   "fake-secret-key",
				BucketName:        "fake-bucket-name",
				CredentialsSource: "static",
				Region:            "fake-region-name",
			}

			expectedS3 := amzs3.New(
				amzaws.Auth{
					AccessKey: "fake-access-key",
					SecretKey: "fake-secret-key",
				},
				config.AWSRegion(),
				amzaws.SignS3,
			)

			client, err := New(config)
			Expect(err).ToNot(HaveOccurred())

			bucket := client.(*amzs3.Bucket)
			Expect(bucket.S3.Auth).To(Equal(expectedS3.Auth))
			Expect(bucket.S3.Region).To(Equal(expectedS3.Region))
			Expect(bucket.S3.Sign).To(BeAssignableToTypeOf(amzaws.SignS3))
			Expect(bucket.Name).To(Equal("fake-bucket-name"))
		})

		It("returns V2 signature", func() {
			config := Config{
				AccessKeyID:       "fake-access-key",
				SecretAccessKey:   "fake-secret-key",
				BucketName:        "fake-bucket-name",
				CredentialsSource: "static",
				Region:            "fake-region-name",
				SignatureVersion:  "2",
			}

			expectedS3 := amzs3.New(
				amzaws.Auth{
					AccessKey: "fake-access-key",
					SecretKey: "fake-secret-key",
				},
				config.AWSRegion(),
				amzaws.SignV2,
			)

			client, err := New(config)
			Expect(err).ToNot(HaveOccurred())

			bucket := client.(*amzs3.Bucket)
			Expect(bucket.S3.Auth).To(Equal(expectedS3.Auth))
			Expect(bucket.S3.Region).To(Equal(expectedS3.Region))
			Expect(bucket.S3.Sign).To(BeAssignableToTypeOf(amzaws.SignV2))
			Expect(bucket.Name).To(Equal("fake-bucket-name"))
		})

		It("returns V4 signature", func() {
			config := Config{
				AccessKeyID:       "fake-access-key",
				SecretAccessKey:   "fake-secret-key",
				BucketName:        "fake-bucket-name",
				CredentialsSource: "static",
				Region:            "fake-region-name",
				SignatureVersion:  "4",
			}

			expectedS3 := amzs3.New(
				amzaws.Auth{
					AccessKey: "fake-access-key",
					SecretKey: "fake-secret-key",
				},
				config.AWSRegion(),
				amzaws.SignV4Factory(config.Region, "s3"),
			)

			client, err := New(config)
			Expect(err).ToNot(HaveOccurred())

			bucket := client.(*amzs3.Bucket)
			Expect(bucket.S3.Auth).To(Equal(expectedS3.Auth))
			Expect(bucket.S3.Region).To(Equal(expectedS3.Region))
			Expect(bucket.S3.Sign).To(BeAssignableToTypeOf(func(*http.Request, amzaws.Auth) error { return nil }))
			Expect(bucket.Name).To(Equal("fake-bucket-name"))
		})
	})
})
