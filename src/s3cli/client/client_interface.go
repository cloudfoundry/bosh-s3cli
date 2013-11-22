package client

import (
	"io"
	amzs3 "launchpad.net/goamz/s3"
)

type S3Client interface {
	PutReader(path string, r io.Reader, length int64, contType string, perm amzs3.ACL) error
}
