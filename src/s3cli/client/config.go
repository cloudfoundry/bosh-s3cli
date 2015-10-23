package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
)

type blobstoreClientConfig struct {
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

const credentialsSourceStatic = "static"
const credentialsSourceEnvOrProfile = "env_or_profile"

func newConfig(file io.Reader) (blobstoreClientConfig, error) {
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return blobstoreClientConfig{}, err
	}

	config := blobstoreClientConfig{
		CredentialsSource: credentialsSourceStatic,
		// Port:              443,
		Region:        "us-east-1",
		SSLVerifyPeer: true,
		UseSSL:        true,
	}

	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return blobstoreClientConfig{}, err
	}

	if config.CredentialsSource == "" {
		config.CredentialsSource = credentialsSourceStatic
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

func (c blobstoreClientConfig) s3Endpoint() string {
	if c.Port != 0 {
		return fmt.Sprintf("%s:%d", c.Host, c.Port)
	}

	return c.Host
}

func (c blobstoreClientConfig) validate() error {
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
