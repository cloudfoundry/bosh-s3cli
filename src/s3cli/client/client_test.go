package client

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAuth(t *testing.T) {
	config, err := getConfig(filepath.Join("../../fixtures/sampleConfig.txt"))
	assert.NoError(t, err)
	assert.Equal(t, "some-access-key", config.AccessKeyID)
	assert.Equal(t, "some-secret-key", config.SecretAccessKey)
	assert.Equal(t, "some-bucket", config.BucketName)
}
