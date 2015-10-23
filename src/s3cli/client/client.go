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

type BlobstoreClient struct {
	s3Client *s3.S3
	config   blobstoreClientConfig
}

func New(configFile io.Reader) (BlobstoreClient, error) {
	config, err := newConfig(configFile)
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
		WithEndpoint(config.s3Endpoint()).
		WithLogLevel(aws.LogOff).
		WithDisableSSL(!config.UseSSL).
		WithHTTPClient(httpClient)

	if config.CredentialsSource == credentialsSourceStatic {
		s3Config = s3Config.WithCredentials(credentials.NewStaticCredentials(config.AccessKeyID, config.SecretAccessKey, ""))
	}

	s3Client := s3.New(s3Config)

	if config.UseV2SigningMethod {
		setv2Handlers(s3Client)
	}

	return BlobstoreClient{s3Client: s3Client, config: config}, nil
}

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

func (c *BlobstoreClient) Put(src io.Reader, dest string) error {
	reader, writer := io.Pipe()

	go func() {
		io.Copy(writer, src)
		writer.Close()
	}()

	uploader := s3manager.NewUploader(&s3manager.UploadOptions{S3: c.s3Client})
	putResult, err := uploader.Upload(&s3manager.UploadInput{
		Body:   reader,
		Bucket: aws.String(c.config.BucketName),
		Key:    aws.String(dest),
	})

	if err != nil {
		return err
	}

	log.Println("Successfully uploaded file to", putResult.Location)
	return nil
}
