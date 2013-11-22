package cmd

import (
	"github.com/stretchr/testify/assert"
	amzs3 "launchpad.net/goamz/s3"
	"os"
	fakeclient "s3cli/client/fakes"
	"testing"
)

func TestPutRun(t *testing.T) {
	client := &fakeclient.FakeClient{}
	factory := NewFactory(client)
	cmd, _ := factory.Create("put")

	err := cmd.Run([]string{"../../fixtures/cat.jpg", "my-cat.jpg"})
	assert.NoError(t, err)

	file := client.PutReaderReader.(*os.File)

	assert.Equal(t, "my-cat.jpg", client.PutReaderPath)
	assert.Equal(t, file.Name(), "../../fixtures/cat.jpg")
	assert.Equal(t, 1718186, client.PutReaderLength)
	assert.Equal(t, "application/octet-stream", client.PutReaderContentType)
	assert.Equal(t, amzs3.BucketOwnerFull, client.PutReaderPerm)
}

func TestPutRunWhenNotEnoughArgument(t *testing.T) {
	client := &fakeclient.FakeClient{}
	factory := NewFactory(client)
	cmd, _ := factory.Create("put")

	err := cmd.Run([]string{"my-cat.jpg"})
	assert.Error(t, err)
}
