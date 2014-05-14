package client

import (
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

	return s3.Bucket(config.BucketName), nil
}
