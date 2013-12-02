package client

import (
	"encoding/json"
	"io/ioutil"
	amzaws "launchpad.net/goamz/aws"
	amzs3 "launchpad.net/goamz/s3"
	"os"
)

type configType struct {
	AccessKey string
	SecretKey string
	Bucket    string
}

func GetS3Client(configPath string) (client S3Client, err error) {
	config, err := getConfig(configPath)
	if err != nil {
		return
	}

	awsAuth := amzaws.Auth{
		AccessKey: config.AccessKey,
		SecretKey: config.SecretKey,
	}

	s3 := amzs3.New(awsAuth, amzaws.USEast)
	client = s3.Bucket(config.Bucket)
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
