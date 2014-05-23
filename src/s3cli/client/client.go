package client

import (
	"crypto/tls"
	"net/http"

	amzaws "launchpad.net/goamz/aws"
	amzs3 "launchpad.net/goamz/s3"
)

func New(config Config) (S3Client, error) {
	s3 := amzs3.New(
		amzaws.Auth{
			AccessKey: config.AccessKeyID,
			SecretKey: config.SecretAccessKey,
		},
		config.AWSRegion(),
	)

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !config.SSLVerifyPeer,
		},
	}

	http.DefaultClient.Transport = transport

	return s3.Bucket(config.BucketName), nil
}
