package fakes

import (
	s3clicmd "s3cli/cmd"
)

type FakeFactory struct {
	CreatedCmdName string
	CreatedCmd     *FakeCmd
}

func (f *FakeFactory) Create(cmdName string) (s3clicmd.Cmd, error) {
	f.CreatedCmdName = cmdName
	return f.CreatedCmd, nil
}

type FakeCmd struct {
	RunArgs []string
}

func (cmd *FakeCmd) Run(args []string) error {
	cmd.RunArgs = args
	return nil
}
