package client_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	amzaws "launchpad.net/goamz/aws"
	amzs3 "launchpad.net/goamz/s3"

	. "s3cli/client"
)

func TestNewClient(t *testing.T) {
	config := Config{
		AccessKeyID:     "fake-access-key",
		SecretAccessKey: "fake-secret-key",
		BucketName:      "fake-bucket-name",
	}
	client, err := New(config)
	assert.NoError(t, err)

	expectedS3 := amzs3.New(
		amzaws.Auth{
			AccessKey: "fake-access-key",
			SecretKey: "fake-secret-key",
		},
		config.AWSRegion(),
	)

	bucket := client.(*amzs3.Bucket)
	assert.Equal(t, expectedS3, bucket.S3)
	assert.Equal(t, "fake-bucket-name", bucket.Name)
}

func TestNewClientWithCustomPort(t *testing.T) {
	config := Config{
		AccessKeyID:     "fake-access-key",
		SecretAccessKey: "fake-secret-key",
		BucketName:      "fake-bucket-name",
		Port:            123,
	}
	client, err := New(config)
	assert.NoError(t, err)

	expectedS3 := amzs3.New(
		amzaws.Auth{
			AccessKey: "fake-access-key",
			SecretKey: "fake-secret-key",
		},
		config.AWSRegion(),
	)

	bucket := client.(*amzs3.Bucket)
	assert.Equal(t, expectedS3, bucket.S3)
	assert.Equal(t, "fake-bucket-name", bucket.Name)
}
