package client_test

import (
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	amzaws "launchpad.net/goamz/aws"

	. "s3cli/client"
)

var (
	expectedDefaultRegion = amzaws.Region{
		Name:        "us-east-1",
		EC2Endpoint: "https://ec2.us-east-1.amazonaws.com",

		S3Endpoint:           "https://s3.amazonaws.com",
		S3BucketEndpoint:     "",
		S3LocationConstraint: false,
		S3LowercaseBucket:    false,

		SDBEndpoint: "https://sdb.amazonaws.com",
		SNSEndpoint: "https://sns.us-east-1.amazonaws.com",
		SQSEndpoint: "https://sqs.us-east-1.amazonaws.com",
		IAMEndpoint: "https://iam.amazonaws.com",
	}
)

var _ = Describe("NewConfigFromPath", func() {
	Context("With minimum configuration", func() {
		It("return config with a default region", func() {
			configPath := writeConfigFile(`{
			  "access_key_id":"some-access-key",
			  "secret_access_key":"some-secret-key",
			  "bucket_name":"some-bucket"
			}`)

			defer os.Remove(configPath)

			config, err := NewConfigFromPath(configPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(config.AccessKeyID).To(Equal("some-access-key"))
			Expect(config.SecretAccessKey).To(Equal("some-secret-key"))
			Expect(config.BucketName).To(Equal("some-bucket"))
			Expect(config.SSLVerifyPeer).To(BeTrue())
			Expect(config.AWSRegion()).To(Equal(expectedDefaultRegion))
		})
	})

	Context("with custom port", func() {
		It("return config with a region with specified port", func() {
			configPath := writeConfigFile(`{
			  "access_key_id":"some-access-key",
			  "secret_access_key":"some-secret-key",
			  "bucket_name":"some-bucket",
			  "port": 123
			}`)

			defer os.Remove(configPath)

			config, err := NewConfigFromPath(configPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(config.AccessKeyID).To(Equal("some-access-key"))
			Expect(config.SecretAccessKey).To(Equal("some-secret-key"))
			Expect(config.BucketName).To(Equal("some-bucket"))

			expectedRegion := expectedDefaultRegion
			expectedRegion.S3Endpoint = "https://s3.amazonaws.com:123"

			Expect(config.AWSRegion()).To(Equal(expectedRegion))
		})
	})

	Context("with custom host but without port", func() {
		It("returns a config with a region with specified host", func() {
			configPath := writeConfigFile(`{
			  "access_key_id":"some-access-key",
			  "secret_access_key":"some-secret-key",
			  "bucket_name":"some-bucket",
			  "host": "host.example.com"
			}`)

			defer os.Remove(configPath)

			config, err := NewConfigFromPath(configPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(config.AccessKeyID).To(Equal("some-access-key"))
			Expect(config.SecretAccessKey).To(Equal("some-secret-key"))
			Expect(config.BucketName).To(Equal("some-bucket"))

			expectedRegion := expectedDefaultRegion
			expectedRegion.S3Endpoint = "https://host.example.com"

			Expect(config.AWSRegion()).To(Equal(expectedRegion))
		})
	})

	Context("with custom host and port", func() {
		It("returns a config with a region with specified host and port", func() {
			configPath := writeConfigFile(`{
			  "access_key_id":"some-access-key",
			  "secret_access_key":"some-secret-key",
			  "bucket_name":"some-bucket",
			  "host": "host.example.com",
			  "port": 123
			}`)

			defer os.Remove(configPath)

			config, err := NewConfigFromPath(configPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(config.AccessKeyID).To(Equal("some-access-key"))
			Expect(config.SecretAccessKey).To(Equal("some-secret-key"))
			Expect(config.BucketName).To(Equal("some-bucket"))

			expectedRegion := expectedDefaultRegion
			expectedRegion.S3Endpoint = "https://host.example.com:123"

			Expect(config.AWSRegion()).To(Equal(expectedRegion))
		})
	})

	Context("with SSL set to true", func() {
		It("returns a config with a region with https scheme", func() {
			configPath := writeConfigFile(`{
			  "access_key_id":"some-access-key",
			  "secret_access_key":"some-secret-key",
			  "bucket_name":"some-bucket",
			  "use_ssl": true
			}`)

			defer os.Remove(configPath)

			config, err := NewConfigFromPath(configPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(config.AccessKeyID).To(Equal("some-access-key"))
			Expect(config.SecretAccessKey).To(Equal("some-secret-key"))
			Expect(config.BucketName).To(Equal("some-bucket"))

			expectedRegion := expectedDefaultRegion
			expectedRegion.S3Endpoint = "https://s3.amazonaws.com"

			Expect(config.AWSRegion()).To(Equal(expectedRegion))
		})
	})

	Context("with SSL set to false", func() {
		It("returns a config with a region with http scheme", func() {
			configPath := writeConfigFile(`{
			  "access_key_id":"some-access-key",
			  "secret_access_key":"some-secret-key",
			  "bucket_name":"some-bucket",
			  "use_ssl": false
			}`)

			defer os.Remove(configPath)

			config, err := NewConfigFromPath(configPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(config.AccessKeyID).To(Equal("some-access-key"))
			Expect(config.SecretAccessKey).To(Equal("some-secret-key"))
			Expect(config.BucketName).To(Equal("some-bucket"))

			expectedRegion := expectedDefaultRegion
			expectedRegion.S3Endpoint = "http://s3.amazonaws.com"

			Expect(config.AWSRegion()).To(Equal(expectedRegion))
		})
	})

	Context("with SSL set to false and custom host and port", func() {
		It("returns a config with a region with https scheme and host and port", func() {
			configPath := writeConfigFile(`{
			  "access_key_id":"some-access-key",
			  "secret_access_key":"some-secret-key",
			  "bucket_name":"some-bucket",
			  "use_ssl": false,
			  "host": "host.example.com",
			  "port": 123
	    }`)

			defer os.Remove(configPath)

			config, err := NewConfigFromPath(configPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(config.AccessKeyID).To(Equal("some-access-key"))
			Expect(config.SecretAccessKey).To(Equal("some-secret-key"))
			Expect(config.BucketName).To(Equal("some-bucket"))

			expectedRegion := expectedDefaultRegion
			expectedRegion.S3Endpoint = "http://host.example.com:123"

			Expect(config.AWSRegion()).To(Equal(expectedRegion))
		})
	})

	Context("with SSL set to false without custom port but with custom host", func() {
		It("returns a config with a region with http scheme and host", func() {
			configPath := writeConfigFile(`{
			  "access_key_id":"some-access-key",
			  "secret_access_key":"some-secret-key",
			  "bucket_name":"some-bucket",
			  "use_ssl": false,
			  "host": "host.example.com"
	    	}`)

			defer os.Remove(configPath)

			config, err := NewConfigFromPath(configPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(config.AccessKeyID).To(Equal("some-access-key"))
			Expect(config.SecretAccessKey).To(Equal("some-secret-key"))
			Expect(config.BucketName).To(Equal("some-bucket"))

			expectedRegion := expectedDefaultRegion
			expectedRegion.S3Endpoint = "http://host.example.com" // no 443

			Expect(config.AWSRegion()).To(Equal(expectedRegion))
		})
	})

	Context("with SSL verify peer set to false", func() {
		It("returns a config with ssl verification set to false", func() {
			configPath := writeConfigFile(`{
				  "access_key_id":"",
			  "ssl_verify_peer": false
	    	}`)

			defer os.Remove(configPath)

			config, err := NewConfigFromPath(configPath)
			Expect(err).ToNot(HaveOccurred())

			Expect(config.SSLVerifyPeer).To(BeFalse())
		})
	})

	Context("with credentials_source set", func() {
		It("returns a correct config", func() {
			configPath := writeConfigFile(`{
			  "credentials_source": "any-source"
			}`)

			defer os.Remove(configPath)

			config, err := NewConfigFromPath(configPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(config.CredentialsSource).To(Equal("any-source"))
		})
	})

	Context("with credentials_source not set", func() {
		It("returns a correct config", func() {
			configPath := writeConfigFile(`{}`)

			defer os.Remove(configPath)

			config, err := NewConfigFromPath(configPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(config.CredentialsSource).To(Equal("static"))
		})
	})

	Context("with credentials_source set to empty string", func() {
		It("returns a correct config", func() {
			configPath := writeConfigFile(`{
			  "credentials_source": ""
			}`)

			defer os.Remove(configPath)

			config, err := NewConfigFromPath(configPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(config.CredentialsSource).To(Equal("static"))
		})
	})
})

func writeConfigFile(contents string) string {
	file, err := ioutil.TempFile("", "client_test")
	Expect(err).ToNot(HaveOccurred())

	err = file.Close()
	Expect(err).ToNot(HaveOccurred())

	err = ioutil.WriteFile(file.Name(), []byte(contents), os.ModeTemporary)
	Expect(err).ToNot(HaveOccurred())

	return file.Name()
}
