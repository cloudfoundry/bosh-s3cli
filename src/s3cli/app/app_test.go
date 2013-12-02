package app

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetConfigPathWithConfigFlag(t *testing.T) {
	path, args, err := getConfigPath([]string{"-c", "/some/path", "foo", "bar"})

	assert.NoError(t, err)
	assert.Equal(t, "/some/path", path)
	assert.Equal(t, []string{"foo", "bar"}, args)
}

func TestGetConfigPathWithoutConfigFlag(t *testing.T) {
	path, args, err := getConfigPath([]string{"foo", "bar"})

	assert.NoError(t, err)
	assert.Contains(t, path, ".s3cli")
	assert.Equal(t, []string{"foo", "bar"}, args)
}
