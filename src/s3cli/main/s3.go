package main

import (
	"fmt"
	"os"
	s3cliapp "s3cli/app"
)

func main() {
	app := s3cliapp.New()
	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}
}
