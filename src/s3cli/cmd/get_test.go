package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	fakeclient "s3cli/client/fakes"
)

func TestGetRun(t *testing.T) {
	fixtureCatPath := "../../fixtures/cat.jpg"
	tmpCatPath := "../../../tmp/cat.jpg"
	fixtureCatFile, err := os.Open(fixtureCatPath)
	assert.NoError(t, err)

	client := &fakeclient.FakeClient{GetReaderReadCloser: fixtureCatFile}
	factory := NewFactory(client)
	cmd, _ := factory.Create("get")

	err = cmd.Run([]string{"my-cat.jpg", tmpCatPath})
	assert.NoError(t, err)
	defer os.RemoveAll(tmpCatPath)

	assert.Equal(t, client.GetReaderPath, "my-cat.jpg")

	tmpCatFile, _ := os.Open(tmpCatPath)
	tmpCatStats, _ := tmpCatFile.Stat()
	assert.Equal(t, tmpCatStats.Size(), 1718186)
}

func TestGetRunWhenNotEnoughArgument(t *testing.T) {
	client := &fakeclient.FakeClient{}
	factory := NewFactory(client)
	cmd, _ := factory.Create("get")

	err := cmd.Run([]string{"my-cat.jpg"})
	assert.Error(t, err)
}
