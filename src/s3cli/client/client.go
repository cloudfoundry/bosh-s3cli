package client

import (
	"encoding/json"
	"io/ioutil"
	amzaws "launchpad.net/goamz/aws"
	amzs3 "launchpad.net/goamz/s3"
	"os"
	"os/user"
	"path/filepath"
)

type configType struct {
	AccessKey string
	SecretKey string
	Bucket    string
}

func GetS3Client() (client S3Client, err error) {
	config, err := getConfig()
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

func getConfig() (config configType, err error) {
	configPath, err := getConfigPath()
	if err != nil {
		return
	}

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

func getConfigPath() (path string, err error) {
	path = os.Getenv("S3_CLI_CONFIG")
	if path != "" {
		return
	}

	usr, err := user.Current()
	if err != nil {
		return
	}

	path = filepath.Join(usr.HomeDir, ".s3cli")
	return
}
