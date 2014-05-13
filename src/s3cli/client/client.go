package client

import (
	amzaws "launchpad.net/goamz/aws"
	amzs3 "launchpad.net/goamz/s3"
)

func New(config Config) (S3Client, error) {
	awsAuth := amzaws.Auth{
		AccessKey: config.AccessKeyID,
		SecretKey: config.SecretAccessKey,
	}

	s3 := amzs3.New(awsAuth, amzaws.USEast)

	return s3.Bucket(config.BucketName), nil
}
