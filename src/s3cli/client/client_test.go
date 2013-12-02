package client

import (
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
)

func TestGetAuth(t *testing.T) {
	configPath := filepath.Join("../../fixtures/sampleConfig.txt")

	config, err := getConfig(configPath)
	assert.NoError(t, err)
	assert.Equal(t, "some-access-key", config.AccessKey)
	assert.Equal(t, "some-access-key", config.AccessKey)
	assert.Equal(t, "some-bucket", config.Bucket)
}
