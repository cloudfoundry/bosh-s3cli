package cmd

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	s3cliclient "s3cli/client"
)

type getCmd struct {
	client s3cliclient.S3Client
}

func newGet(s3Client s3cliclient.S3Client) (cmd Cmd) {
	return getCmd{
		client: s3Client,
	}
}

func (cmd getCmd) Run(args []string) (err error) {
	if len(args) < 2 {
		err = errors.New("Not enough arguments, expected remote path and destination path")
		return
	}

	remotePath := args[0]
	localPath := args[1]

	readCloser, err := cmd.client.GetReader(remotePath)
	if err != nil {
		return
	}
	defer readCloser.Close()

	err = os.MkdirAll(filepath.Dir(localPath), os.ModePerm)
	if err != nil {
		return
	}

	targetFile, err := os.Create(localPath)
	if err != nil {
		return
	}

	_, err = io.Copy(targetFile, readCloser)
	return
}
