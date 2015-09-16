package cmd

import (
	"errors"
	"os"

	amzs3 "gopkg.in/amz.v3/s3"
	s3cliclient "s3cli/client"
)

type putCmd struct {
	client s3cliclient.S3Client
}

func newPut(s3Client s3cliclient.S3Client) Cmd {
	return putCmd{client: s3Client}
}

func (cmd putCmd) Run(args []string) (err error) {
	if len(args) < 2 {
		return errors.New("Not enough arguments, expected source file and destination path")
	}

	source := args[0]
	destination := args[1]

	file, err := os.Open(source)
	if err != nil {
		return err
	}

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	return cmd.client.PutReader(
		destination,
		file,
		stat.Size(),
		"application/octet-stream",
		amzs3.BucketOwnerFull,
	)
}
