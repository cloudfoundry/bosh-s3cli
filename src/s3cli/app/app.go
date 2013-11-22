package app

import (
	"errors"
	s3clicmd "s3cli/cmd"
)

type app struct {
	cmdRunner s3clicmd.Runner
}

func New(cmdRunner s3clicmd.Runner) (app app) {
	app.cmdRunner = cmdRunner
	return
}

func (app app) Run(args []string) (err error) {
	if len(args) < 2 {
		err = errors.New("Command missing")
		return
	}

	cmdName := args[1]
	cmdArgs := args[2:]
	err = app.cmdRunner.Run(cmdName, cmdArgs)
	return
}
