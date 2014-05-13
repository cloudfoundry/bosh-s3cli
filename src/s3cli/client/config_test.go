package client_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	. "s3cli/client"
)

func TestNewConfigFromPath(t *testing.T) {
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
