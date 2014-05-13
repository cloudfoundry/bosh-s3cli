package client

import (
	"encoding/json"
	"io/ioutil"
	"os"

	amzaws "launchpad.net/goamz/aws"
	amzs3 "launchpad.net/goamz/s3"
)

type configType struct {
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	BucketName      string `json:"bucket_name"`
}

func GetS3Client(configPath string) (client S3Client, err error) {
	config, err := getConfig(configPath)
	if err != nil {
		return
	}

	awsAuth := amzaws.Auth{
		AccessKey: config.AccessKeyID,
		SecretKey: config.SecretAccessKey,
	}

	s3 := amzs3.New(awsAuth, amzaws.USEast)
	client = s3.Bucket(config.BucketName)
	return
}

func getConfig(configPath string) (config configType, err error) {
	file, err := os.Open(configPath)
	if err != nil {
		return
	}

	configBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return
	}

	err = json.Unmarshal(configBytes, &config)
	return
}
