package main

import (
	"fmt"
	"os"
	s3cliapp "s3cli/app"
	s3cliclient "s3cli/client"
	s3clicmd "s3cli/cmd"
)

func main() {
	s3Client, err := s3cliclient.GetS3Client()
	handlerError(err)

	cmdFactory := s3clicmd.NewFactory(s3Client)
	cmdRunner := s3clicmd.NewRunner(cmdFactory)
	app := s3cliapp.New(cmdRunner)

	err = app.Run(os.Args)
	handlerError(err)
}

func handlerError(err error) {
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}
}
