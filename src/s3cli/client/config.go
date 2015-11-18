package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
)

type BlobstoreClientConfig struct {
	AccessKeyID        string `json:"access_key_id"`
	SecretAccessKey    string `json:"secret_access_key"`
	BucketName         string `json:"bucket_name"`
	CredentialsSource  string `json:"credentials_source"`
	Host               string `json:"host"`
	Port               int    `json:"port"` // 0 means no custom port
	Region             string `json:"region"`
	SSLVerifyPeer      bool   `json:"ssl_verify_peer"`
	UseSSL             bool   `json:"use_ssl"`
	UseV2SigningMethod bool
}

const (
	credentialsSourceStatic = "static"
	credentialsSourceEnvOrProfile = "env_or_profile"
	defaultRegion = "us-east-1"
)

func newConfig(file io.Reader) (BlobstoreClientConfig, error) {
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return BlobstoreClientConfig{}, err
	}

	config := BlobstoreClientConfig{
		SSLVerifyPeer:     true,
		UseSSL:            true,
	}

	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return BlobstoreClientConfig{}, err
	}

	if config.CredentialsSource == "" {
		config.CredentialsSource = credentialsSourceStatic
	}

	if config.Region == "" && config.Host == "" {
		config.Region = defaultRegion
	}

	switch config.Region {
	case "eu-central-1":
		// use v4 signing
	case "cn-north-1":
		// use v4 signing
	default:
		config.UseV2SigningMethod = true
	}

	return config, nil
}

func (c BlobstoreClientConfig) s3Endpoint() string {
	if c.Port != 0 {
		return fmt.Sprintf("%s:%d", c.Host, c.Port)
	}

	return c.Host
}

func (c BlobstoreClientConfig) validate() error {
	switch c.CredentialsSource {
	case credentialsSourceStatic:
		if c.AccessKeyID == "" || c.SecretAccessKey == "" {
			return errorStaticCredentialsMissing
		}
	case credentialsSourceEnvOrProfile:
		if c.AccessKeyID != "" || c.SecretAccessKey != "" {
			return errorStaticCredentialsPresent
		}
	default:
		return fmt.Errorf("Incorrect credentials_source: %s", c.CredentialsSource)
	}

	return nil
}

var errorStaticCredentialsMissing = errors.New("BlobstoreClient: access_key_id and secret_access_key must be provided")
var errorStaticCredentialsPresent = errors.New("BlobstoreClient: Can't use access_key_id and secret_access_key with env_or_profile credentials_source")
