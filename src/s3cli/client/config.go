package client

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strconv"

	amzaws "launchpad.net/goamz/aws"
)

type Config struct {
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	BucketName      string `json:"bucket_name"`

	Host string `json:"host"`
	Port int    `json:"port"` // 0 means no custom port
}

func NewConfigFromPath(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}

	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return Config{}, err
	}

	var config Config

	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}

func (c Config) AWSRegion() amzaws.Region {
	host := "s3.amazonaws.com"
	if c.Host != "" {
		host = c.Host
	}

	s3Endpoint := "https://" + host

	if c.Port != 0 {
		s3Endpoint += ":" + strconv.Itoa(c.Port)
	}

	return amzaws.Region{
		Name:        "us-east-1",
		EC2Endpoint: "https://ec2.us-east-1.amazonaws.com",

		S3Endpoint:           s3Endpoint,
		S3BucketEndpoint:     "",
		S3LocationConstraint: false,
		S3LowercaseBucket:    false,

		SDBEndpoint: "https://sdb.amazonaws.com",
		SNSEndpoint: "https://sns.us-east-1.amazonaws.com",
		SQSEndpoint: "https://sqs.us-east-1.amazonaws.com",
		IAMEndpoint: "https://iam.amazonaws.com",
	}
}
