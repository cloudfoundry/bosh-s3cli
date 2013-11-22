package client

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestGetAuth(t *testing.T) {
	os.Setenv("S3_CLI_CONFIG", filepath.Join("../../fixtures/sampleConfig.txt"))

	config, err := getConfig()
	assert.NoError(t, err)
	assert.Equal(t, "some-access-key", config.AccessKey)
	assert.Equal(t, "some-access-key", config.AccessKey)
	assert.Equal(t, "some-bucket", config.Bucket)
}
