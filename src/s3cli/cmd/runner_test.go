package cmd

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type FakeFactory struct {
	CreatedCmdName string
	CreatedCmd     *FakeCmd
}

func (factory *FakeFactory) Create(cmdName string) (cmd Cmd, err error) {
	factory.CreatedCmdName = cmdName
	cmd = factory.CreatedCmd
	return
}

type FakeCmd struct {
	RunArgs []string
}

func (cmd *FakeCmd) Run(args []string) (err error) {
	cmd.RunArgs = args
	return
}

func TestRunnerRun(t *testing.T) {
	fakeFactory := &FakeFactory{
		CreatedCmd: &FakeCmd{},
	}

	runner := NewRunner(fakeFactory)
	err := runner.Run("some-cmd", []string{"param1", "param2"})

	assert.NoError(t, err)
	assert.Equal(t, "some-cmd", fakeFactory.CreatedCmdName)
	assert.Equal(t, []string{"param1", "param2"}, fakeFactory.CreatedCmd.RunArgs)
}
