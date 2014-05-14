package client_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestNewConfigFromPathWithDefaultPort(t *testing.T) {
	configPath := writeConfigFile(t, `{
	  "access_key_id":"some-access-key",
	  "secret_access_key":"some-secret-key",
	  "bucket_name":"some-bucket"
	}`)

	defer os.Remove(configPath)

	config, err := NewConfigFromPath(configPath)
	assert.NoError(t, err)
	assert.Equal(t, "some-access-key", config.AccessKeyID)
	assert.Equal(t, "some-secret-key", config.SecretAccessKey)
	assert.Equal(t, "some-bucket", config.BucketName)
	assert.Equal(t, expectedDefaultRegion, config.AWSRegion())
}

func TestNewConfigFromPathWithPort(t *testing.T) {
	configPath := writeConfigFile(t, `{
	  "access_key_id":"some-access-key",
	  "secret_access_key":"some-secret-key",
	  "bucket_name":"some-bucket",
	  "port": 123
	}`)

	defer os.Remove(configPath)

	config, err := NewConfigFromPath(configPath)
	assert.NoError(t, err)
	assert.Equal(t, "some-access-key", config.AccessKeyID)
	assert.Equal(t, "some-secret-key", config.SecretAccessKey)
	assert.Equal(t, "some-bucket", config.BucketName)

	expectedRegion := expectedDefaultRegion
	expectedRegion.S3Endpoint = "https://s3.amazonaws.com:123"

	assert.Equal(t, expectedRegion, config.AWSRegion())
}

func writeConfigFile(t *testing.T, contents string) string {
	file, err := ioutil.TempFile("", "client_test")
	assert.NoError(t, err)

	err = file.Close()
	assert.NoError(t, err)

	err = ioutil.WriteFile(file.Name(), []byte(contents), os.ModeTemporary)
	assert.NoError(t, err)

	return file.Name()
}
