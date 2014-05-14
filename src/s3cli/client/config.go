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

	Host   string `json:"host"`
	Port   int    `json:"port"` // 0 means no custom port
	UseSSL bool   `json:"use_ssl"`
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

	config := Config{UseSSL: true, Port: 443}

	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}

func (c Config) AWSRegion() amzaws.Region {
	return amzaws.Region{
		Name:        "us-east-1",
		EC2Endpoint: "https://ec2.us-east-1.amazonaws.com",

		S3Endpoint:           c.s3Endpoint(),
		S3BucketEndpoint:     "",
		S3LocationConstraint: false,
		S3LowercaseBucket:    false,

		SDBEndpoint: "https://sdb.amazonaws.com",
		SNSEndpoint: "https://sns.us-east-1.amazonaws.com",
		SQSEndpoint: "https://sqs.us-east-1.amazonaws.com",
		IAMEndpoint: "https://iam.amazonaws.com",
	}
}

func (c Config) s3Endpoint() string {
	host := "s3.amazonaws.com"
	if c.Host != "" {
		host = c.Host
	}

	scheme := "https"
	if !c.UseSSL {
		scheme = "http"
	}

	portSuffix := ""
	if c.Port != 443 {
		portSuffix = ":" + strconv.Itoa(c.Port)
	}

	return scheme + "://" + host + portSuffix
}
