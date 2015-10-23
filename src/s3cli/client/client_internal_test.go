package client

import (
	"bytes"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BlobstoreClient configuration", func() {
	Describe("building a configuration", func() {
		Describe("default arguments", func() {
			It("has sensible defaults for all the properties", func() {
				emptyJSONBytes := []byte(`{}`)
				emptyJSONReader := bytes.NewReader(emptyJSONBytes)

				config, err := newConfig(emptyJSONReader)
				Expect(err).ToNot(HaveOccurred())
				Expect(config.BucketName).To(Equal(""))
				Expect(config.CredentialsSource).To(Equal("static"))
				Expect(config.Host).To(Equal(""))
				Expect(config.Port).To(Equal(0))
				Expect(config.Region).To(Equal("us-east-1"))
				Expect(config.SSLVerifyPeer).To(BeTrue())
				Expect(config.UseSSL).To(BeTrue())
				Expect(config.UseV2SigningMethod).To(BeTrue())
			})
		})

		It("sets `credentials_source` to `static` if an empty string is provided in the configuration", func() {
			dummyJSONBytes := []byte(`{"credentials_source": ""}`)
			dummyJSONReader := bytes.NewReader(dummyJSONBytes)

			config, err := newConfig(dummyJSONReader)
			Expect(err).ToNot(HaveOccurred())
			Expect(config.CredentialsSource).To(Equal("static"))
		})

		Describe("configuring v4 signing", func() {
			It("uses v4 signing in the `eu-central-1` AWS region", func() {
				dummyJSONBytes := []byte(`{"region": "eu-central-1"}`)
				dummyJSONReader := bytes.NewReader(dummyJSONBytes)

				config, err := newConfig(dummyJSONReader)
				Expect(err).ToNot(HaveOccurred())
				Expect(config.UseV2SigningMethod).To(BeFalse())
			})

			It("uses v4 signing in the `cn-north-1` AWS region", func() {
				dummyJSONBytes := []byte(`{"region": "cn-north-1"}`)
				dummyJSONReader := bytes.NewReader(dummyJSONBytes)

				config, err := newConfig(dummyJSONReader)
				Expect(err).ToNot(HaveOccurred())
				Expect(config.UseV2SigningMethod).To(BeFalse())
			})
		})

		Context("when the configuration file cannot be read", func() {
			It("returns an error", func() {
				explodingReader := explodeOnRead{}

				_, err := newConfig(explodingReader)
				Expect(err).To(MatchError("explosion"))
			})
		})

		Context("when the configuration file is invalid JSON", func() {
			It("returns an error", func() {
				invalidJSONBytes := []byte(`invalid-json`)
				invalidJSONReader := bytes.NewReader(invalidJSONBytes)

				_, err := newConfig(invalidJSONReader)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("returning the S3 endpoint", func() {
		Context("when port is provided", func() {
			It("returns a URI in the form host:port", func() {
				dummyJSONBytes := []byte(`{"host": "some-host-name", "port": 443}`)
				dummyJSONReader := bytes.NewReader(dummyJSONBytes)

				config, err := newConfig(dummyJSONReader)
				Expect(err).ToNot(HaveOccurred())
				Expect(config.s3Endpoint()).To(Equal("some-host-name:443"))
			})
		})

		Context("when port is omitted", func() {
			It("returns the host", func() {
				dummyJSONBytes := []byte(`{"host": "some-host-name"}`)
				dummyJSONReader := bytes.NewReader(dummyJSONBytes)

				config, err := newConfig(dummyJSONReader)
				Expect(err).ToNot(HaveOccurred())
				Expect(config.s3Endpoint()).To(Equal("some-host-name"))
			})
		})
	})

	Describe("validating credentials", func() {
		Context("when credential source is `static`", func() {
			It("validates that access key and secret key are set", func() {
				dummyJSONBytes := []byte(`{}`)
				dummyJSONReader := bytes.NewReader(dummyJSONBytes)

				config, err := newConfig(dummyJSONReader)
				Expect(err).ToNot(HaveOccurred())
				err = config.validate()
				Expect(err).To(MatchError(errorStaticCredentialsMissing))

				dummyJSONBytes = []byte(`{"access_key_id": "some_id"}`)
				dummyJSONReader = bytes.NewReader(dummyJSONBytes)
				config, err = newConfig(dummyJSONReader)
				Expect(err).ToNot(HaveOccurred())
				err = config.validate()
				Expect(err).To(MatchError(errorStaticCredentialsMissing))

				dummyJSONBytes = []byte(`{"access_key_id": "some_id", "secret_access_key": "some_secret"}`)
				dummyJSONReader = bytes.NewReader(dummyJSONBytes)
				config, err = newConfig(dummyJSONReader)
				Expect(err).ToNot(HaveOccurred())
				err = config.validate()
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when credentials source is `env_or_profile`", func() {
			It("validates that access key and secret key are not set", func() {
				dummyJSONBytes := []byte(`{"credentials_source": "env_or_profile"}`)
				dummyJSONReader := bytes.NewReader(dummyJSONBytes)

				config, err := newConfig(dummyJSONReader)
				Expect(err).ToNot(HaveOccurred())
				err = config.validate()
				Expect(err).ToNot(HaveOccurred())

				dummyJSONBytes = []byte(`{"credentials_source": "env_or_profile", "access_key_id": "some_id"}`)
				dummyJSONReader = bytes.NewReader(dummyJSONBytes)
				config, err = newConfig(dummyJSONReader)
				Expect(err).ToNot(HaveOccurred())
				err = config.validate()
				Expect(err).To(MatchError(errorStaticCredentialsPresent))

				dummyJSONBytes = []byte(`{"credentials_source": "env_or_profile", "access_key_id": "some_id", "secret_access_key": "some_secret"}`)
				dummyJSONReader = bytes.NewReader(dummyJSONBytes)
				config, err = newConfig(dummyJSONReader)
				Expect(err).ToNot(HaveOccurred())
				err = config.validate()
				Expect(err).To(MatchError(errorStaticCredentialsPresent))
			})
		})

		Context("when credentials source is invalid", func() {
			It("returns an error", func() {
				dummyJSONBytes := []byte(`{"credentials_source": "magical_unicorns"}`)
				dummyJSONReader := bytes.NewReader(dummyJSONBytes)

				config, err := newConfig(dummyJSONReader)
				Expect(err).ToNot(HaveOccurred())
				err = config.validate()
				Expect(err).To(MatchError("Incorrect credentials_source: magical_unicorns"))
			})
		})
	})
})

type explodeOnRead struct{}

func (e explodeOnRead) Read([]byte) (int, error) {
	return 0, errors.New("explosion")
}
