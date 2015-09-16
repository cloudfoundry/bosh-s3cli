package client

import (
	"io"

	amzs3 "gopkg.in/amz.v3/s3"
)

type S3Client interface {
	GetReader(path string) (rc io.ReadCloser, err error)
	PutReader(path string, r io.Reader, length int64, contType string, perm amzs3.ACL) error
}
