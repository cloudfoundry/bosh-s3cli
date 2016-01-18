package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
)

// The S3Cli represents configuration for the s3cli
type S3Cli struct {
	AccessKeyID        string `json:"access_key_id"`
	SecretAccessKey    string `json:"secret_access_key"`
	BucketName         string `json:"bucket_name"`
	CredentialsSource  string `json:"credentials_source"`
	Host               string `json:"host"`
	Port               int    `json:"port"` // 0 means no custom port
	Region             string `json:"region"`
	SSLVerifyPeer      bool   `json:"ssl_verify_peer"`
	UseSSL             bool   `json:"use_ssl"`
	SignatureVersion   int    `json:"signature_version,string"`
	UseV2SigningMethod bool
}

// EmptyRegion is required to allow us to use the AWS SDK against S3 compatible blobstores which do not have
// the concept of a region
const EmptyRegion = " "

// StaticCredentialsSource specifies that credentials will be supplied using access_key_id and secret_access_key
const StaticCredentialsSource = "static"
// NoneCredentialsSource specifies that credentials will be empty. The blobstore client operates in read only mode.
const NoneCredentialsSource = "none"

const (
	credentialsSourceEnvOrProfile = "env_or_profile"
	defaultRegion                 = "us-east-1"
	euCentralRegion               = "eu-central-1"
	cnNorthRegion                 = "cn-north-1"
)

var errorStaticCredentialsMissing = errors.New("access_key_id and secret_access_key must be provided")
var errorStaticCredentialsPresent = errors.New("can't use access_key_id and secret_access_key with env_or_profile credentials_source")

// NewFromReader returns a new s3cli configuration struct from the contents of reader.
// reader.Read() is expected to return valid JSON
func NewFromReader(reader io.Reader) (S3Cli, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return S3Cli{}, err
	}

	c := S3Cli{
		SSLVerifyPeer: true,
		UseSSL:        true,
	}

	err = json.Unmarshal(bytes, &c)
	if err != nil {
		return S3Cli{}, err
	}

	if c.BucketName == "" {
		return S3Cli{}, errors.New("bucket_name must be set")
	}

	if c.CredentialsSource == "" {
		c.CredentialsSource = StaticCredentialsSource
	}

	switch c.CredentialsSource {
	case StaticCredentialsSource:
		if c.AccessKeyID == "" || c.SecretAccessKey == "" {
			return S3Cli{}, errorStaticCredentialsMissing
		}
	case credentialsSourceEnvOrProfile:
		if c.AccessKeyID != "" || c.SecretAccessKey != "" {
			return S3Cli{}, errorStaticCredentialsPresent
		}
	case NoneCredentialsSource:
	default:
		return S3Cli{}, fmt.Errorf("Invalid credentials_source: %s", c.CredentialsSource)
	}

	if c.Region == "" && c.Host == "" {
		c.Region = defaultRegion
	}

	switch c.SignatureVersion {
	case 2:
		c.UseV2SigningMethod = true
	case 4:
		c.UseV2SigningMethod = false
	default:
		if c.Region == "" && c.Host != "" {
			isHostAWS, _ := regexp.MatchString("amazonaws", c.Host)
			if !isHostAWS || c.CredentialsSource == StaticCredentialsSource {
				c.UseV2SigningMethod = true
			}
		}
	}

	return c, nil
}

// UseRegion signals to users of the S3Cli whether to use Region information
func (c *S3Cli) UseRegion() bool {
	return c.Region != ""
}

// S3Endpoint returns the S3 URI to use if custom host information has been provided
func (c *S3Cli) S3Endpoint() string {
	if c.Host == "" {
		return ""
	}
	if c.Port != 0 {
		return fmt.Sprintf("%s:%d", c.Host, c.Port)
	}
	return c.Host
}
