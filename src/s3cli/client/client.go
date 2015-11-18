package client

import (
	"crypto/tls"
	"io"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// BlobstoreClient encapsulates interacting with an S3 compatible blobstore
type BlobstoreClient struct {
	s3Client *s3.S3
	config   BlobstoreClientConfig
}

// New returns a BlobstoreClient if the configuration file backing configFile is valid
func New(configFile io.Reader) (BlobstoreClient, error) {
	config, err := newConfig(configFile)
	if err != nil {
		return BlobstoreClient{}, err
	}

	err = config.validate()
	if err != nil {
		return BlobstoreClient{}, err
	}

	transport := *http.DefaultTransport.(*http.Transport)
	transport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: !config.SSLVerifyPeer,
	}

	httpClient := &http.Client{Transport: &transport}

	s3Config := aws.NewConfig().
		WithRegion(config.Region).
		WithS3ForcePathStyle(true).
		WithLogLevel(aws.LogOff).
		WithDisableSSL(!config.UseSSL).
		WithHTTPClient(httpClient)

	if config.Region != "" && config.Host == "" {
		s3Config = s3Config.WithRegion(config.Region)
	} else if config.Host != "" && config.Region == "" {
		s3Config = s3Config.WithEndpoint(config.s3Endpoint())
	} else {
		panic("unable to handle specifying both host and region")
	}

	if config.CredentialsSource == credentialsSourceStatic {
		s3Config = s3Config.WithCredentials(credentials.NewStaticCredentials(config.AccessKeyID, config.SecretAccessKey, ""))
	}

	s3Client := s3.New(s3Config)

	if config.UseV2SigningMethod {
		setv2Handlers(s3Client)
	}

	return BlobstoreClient{s3Client: s3Client, config: config}, nil
}

// Get fetches a blob from an S3 compatible blobstore
// Destination will be overwritten if exists
func (c *BlobstoreClient) Get(src string, dest io.WriterAt) error {
	downloader := s3manager.NewDownloader(&s3manager.DownloadOptions{S3: c.s3Client})

	_, err := downloader.Download(dest, &s3.GetObjectInput{
		Bucket: aws.String(c.config.BucketName),
		Key:    aws.String(src),
	})

	if err != nil {
		return err
	}

	return nil
}

// Put uploads a blob to an S3 compatible blobstore
func (c *BlobstoreClient) Put(src io.ReadSeeker, dest string) error {
	uploader := s3manager.NewUploader(&s3manager.UploadOptions{S3: c.s3Client})
	putResult, err := uploader.Upload(&s3manager.UploadInput{
		Body:   src,
		Bucket: aws.String(c.config.BucketName),
		Key:    aws.String(dest),
	})

	if err != nil {
		return err
	}

	log.Println("Successfully uploaded file to", putResult.Location)
	return nil
}
