package app

import (
	"errors"
	"flag"
	"io/ioutil"
	"os/user"
	"path/filepath"
	s3cliclient "s3cli/client"
	s3clicmd "s3cli/cmd"
)

type app struct {
}

func New() (app app) {
	return
}

func (app app) Run(args []string) (err error) {
	args = args[1:]

	configPath, args, err := getConfigPath(args)
	if err != nil {
		return
	}

	s3Client, err := s3cliclient.GetS3Client(configPath)
	if err != nil {
		return
	}

	cmdFactory := s3clicmd.NewFactory(s3Client)
	cmdRunner := s3clicmd.NewRunner(cmdFactory)

	if len(args) < 1 {
		err = errors.New("Command missing")
		return
	}

	cmdName := args[0]
	cmdArgs := args[1:]
	err = cmdRunner.Run(cmdName, cmdArgs)
	return
}

func getConfigPath(args []string) (path string, updatedArgs []string, err error) {
	flagSet := flag.NewFlagSet("s3-cli-args", flag.ContinueOnError)
	flagSet.SetOutput(ioutil.Discard)
	flagSet.StringVar(&path, "c", "", "Config file")
	flagSet.Parse(args)

	if path != "" {
		updatedArgs = args[2:]
		return
	}

	updatedArgs = args

	usr, err := user.Current()
	if err != nil {
		return
	}

	path = filepath.Join(usr.HomeDir, ".s3cli")
	return
}
