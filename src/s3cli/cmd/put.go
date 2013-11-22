package cmd

import (
	"errors"
	amzs3 "launchpad.net/goamz/s3"
	"os"
	s3cliclient "s3cli/client"
)

type putCmd struct {
	client s3cliclient.S3Client
}

func newPut(s3Client s3cliclient.S3Client) (cmd Cmd) {
	return putCmd{
		client: s3Client,
	}
}

func (cmd putCmd) Run(args []string) (err error) {
	if len(args) < 2 {
		err = errors.New("Not enough arguments, expected source file and destination path")
		return
	}

	source := args[0]
	destination := args[1]

	file, err := os.Open(source)
	if err != nil {
		return
	}

	stat, err := file.Stat()
	if err != nil {
		return
	}

	err = cmd.client.PutReader(
		destination,
		file,
		stat.Size(),
		"application/octet-stream",
		amzs3.BucketOwnerFull,
	)
	return
}
