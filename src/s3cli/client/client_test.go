package client_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	amzaws "launchpad.net/goamz/aws"
	amzs3 "launchpad.net/goamz/s3"

	. "s3cli/client"
)

type HTTPHandler func(http.ResponseWriter, *http.Request)

func (h HTTPHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) { h(rw, r) }

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
				AccessKeyID:     "fake-access-key",
				SecretAccessKey: "fake-secret-key",
				BucketName:      "fake-bucket-name",
				Host:            hostName,
				Port:            port,
				UseSSL:          true,
				SSLVerifyPeer:   true,
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
})
