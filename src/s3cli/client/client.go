package client

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"

	amzaws "gopkg.in/amz.v3/aws"
	amzs3 "gopkg.in/amz.v3/s3"
)

func New(config Config) (S3Client, error) {
	awsAuth := amzaws.Auth{
		AccessKey: config.AccessKeyID,
		SecretKey: config.SecretAccessKey,
	}

	switch config.CredentialsSource {
	case "static":
		if config.AccessKeyID == "" || config.SecretAccessKey == "" {
			return nil, errors.New("access_key_id or secret_access_key is missing")
		}
	case "env_or_profile":
		if config.AccessKeyID == "" && config.SecretAccessKey == "" {
			auth, err := amzaws.GetAuth()
			if err != nil {
				return nil, err
			}
			awsAuth = amzaws.NewAuth(auth.AccessKey, auth.SecretKey, auth.Token(), auth.Expiration())
		} else {
			return nil, errors.New("Can't use access_key_id and secret_access_key with env_or_profile credentials_source")
		}
	default:
		return nil, fmt.Errorf("Incorrect credentials_source: %s", config.CredentialsSource)
	}

	var signer amzaws.Signer
	switch config.SignatureVersion {
	case "4":
		signer = amzaws.SignV4Factory(config.Region, "s3")
	case "2":
		signer = amzaws.SignV2
	default:
		signer = amzaws.SignS3
	}

	s3 := amzs3.New(awsAuth, config.AWSRegion(), signer)

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !config.SSLVerifyPeer,
		},
	}

	http.DefaultClient.Transport = transport

	return s3.Bucket(config.BucketName)
}
