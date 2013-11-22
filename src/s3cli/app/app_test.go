package app

import (
	"github.com/stretchr/testify/assert"
	fakecmd "s3cli/cmd/fakes"
	"testing"
)

func TestRun(t *testing.T) {
	fakeRunner := &fakecmd.FakeRunner{}
	app := New(fakeRunner)
	err := app.Run([]string{"s3", "put", "/tmp/file.txt", "new-name.txt"})

	assert.NoError(t, err)
	assert.Equal(t, "put", fakeRunner.RunCmdName)
	assert.Equal(t, []string{"/tmp/file.txt", "new-name.txt"}, fakeRunner.RunCmdArgs)
}

func TestRunWithoutCommand(t *testing.T) {
	fakeRunner := &fakecmd.FakeRunner{}
	app := New(fakeRunner)
	err := app.Run([]string{"s3"})

	assert.Error(t, err)
}

func TestRunWithOnlyOneArgument(t *testing.T) {
	fakeRunner := &fakecmd.FakeRunner{}
	app := New(fakeRunner)
	err := app.Run([]string{"s3", "put"})

	assert.NoError(t, err)
	assert.Equal(t, "put", fakeRunner.RunCmdName)
	assert.Equal(t, []string{}, fakeRunner.RunCmdArgs)
}
