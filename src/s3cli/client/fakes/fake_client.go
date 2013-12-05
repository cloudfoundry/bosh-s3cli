package fakes

import (
	"io"
	amzs3 "launchpad.net/goamz/s3"
)

type FakeClient struct {
	GetReaderReadCloser io.ReadCloser
	GetReaderError      error
	GetReaderPath       string

	PutReaderPath        string
	PutReaderReader      io.Reader
	PutReaderLength      int64
	PutReaderContentType string
	PutReaderPerm        amzs3.ACL
}

func (client *FakeClient) GetReader(path string) (rc io.ReadCloser, err error) {
	rc = client.GetReaderReadCloser
	err = client.GetReaderError
	client.GetReaderPath = path
	return
}

func (client *FakeClient) PutReader(path string, r io.Reader, length int64, contType string, perm amzs3.ACL) error {
	client.PutReaderPath = path
	client.PutReaderReader = r
	client.PutReaderLength = length
	client.PutReaderContentType = contType
	client.PutReaderPerm = perm
	return nil
}
