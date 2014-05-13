package client

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type Config struct {
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	BucketName      string `json:"bucket_name"`
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
