package client

import (
	"crypto/tls"
	"errors"
	"io"
	"log"
	"net/http"

	"s3cli/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// S3Blobstore encapsulates interactions with an S3 compatible blobstore
type S3Blobstore struct {
	s3Client    *s3.S3
	s3cliConfig config.S3Cli
}

var errorInvalidCredentialsSourceValue = errors.New("the client operates in read only mode. Change 'credentials_source' parameter value ")

// New returns a BlobstoreClient if the configuration file backing configFile is valid
func New(configFile io.Reader) (S3Blobstore, error) {
	c, err := config.NewFromReader(configFile)
	if err != nil {
		return S3Blobstore{}, err
	}

	transport := *http.DefaultTransport.(*http.Transport)
	transport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: !c.SSLVerifyPeer,
	}

	httpClient := &http.Client{Transport: &transport}

	s3Config := aws.NewConfig().
		WithLogLevel(aws.LogOff).
		WithS3ForcePathStyle(true).
		WithDisableSSL(!c.UseSSL).
		WithHTTPClient(httpClient)

	if c.UseRegion() {
		s3Config = s3Config.WithRegion(c.Region).WithEndpoint(c.S3Endpoint())
	} else {
		s3Config = s3Config.WithRegion(config.EmptyRegion).WithEndpoint(c.S3Endpoint())
	}

	if c.CredentialsSource == config.StaticCredentialsSource {
		s3Config = s3Config.WithCredentials(credentials.NewStaticCredentials(c.AccessKeyID, c.SecretAccessKey, ""))
	}

	if c.CredentialsSource == config.NoneCredentialsSource {
		s3Config = s3Config.WithCredentials(credentials.AnonymousCredentials)
	}

	s3Session := session.New(s3Config)
	s3Client := s3.New(s3Session)

	if c.UseV2SigningMethod {
		setv2Handlers(s3Client)
	}

	return S3Blobstore{s3Client: s3Client, s3cliConfig: c}, nil
}

// Get fetches a blob from an S3 compatible blobstore
// Destination will be overwritten if exists
func (client *S3Blobstore) Get(src string, dest io.WriterAt) error {
	downloader := s3manager.NewDownloaderWithClient(client.s3Client)

	_, err := downloader.Download(dest, &s3.GetObjectInput{
		Bucket: aws.String(client.s3cliConfig.BucketName),
		Key:    aws.String(src),
	})

	if err != nil {
		return err
	}

	return nil
}

// Put uploads a blob to an S3 compatible blobstore
func (client *S3Blobstore) Put(src io.ReadSeeker, dest string) error {
	if client.s3cliConfig.CredentialsSource == config.NoneCredentialsSource {
		return errorInvalidCredentialsSourceValue
	}

	uploader := s3manager.NewUploaderWithClient(client.s3Client)
	putResult, err := uploader.Upload(&s3manager.UploadInput{
		Body:   src,
		Bucket: aws.String(client.s3cliConfig.BucketName),
		Key:    aws.String(dest),
	})

	if err != nil {
		return err
	}

	log.Println("Successfully uploaded file to", putResult.Location)
	return nil
}

// Delete remove a blob from an S3 compatible blobstore
func (client *S3Blobstore) Delete(dest string) error {
	if client.s3cliConfig.CredentialsSource == config.NoneCredentialsSource {
		return errorInvalidCredentialsSourceValue
	}

	deleteParams := &s3.DeleteObjectInput{
		Bucket: aws.String(client.s3cliConfig.BucketName),
		Key:    aws.String(dest),
	}

	_, err := client.s3Client.DeleteObject(deleteParams)

	if err != nil {
		return err
	}

	return nil
}

// Exist checks if blob exists in an S3 compatible blobstore
func (client *S3Blobstore) Exist(dest string) (bool, error) {

	existsParams := &s3.HeadObjectInput{
		Bucket: aws.String(client.s3cliConfig.BucketName),
		Key:    aws.String(dest),
	}

	_, err := client.s3Client.HeadObject(existsParams)

	if err != nil {
		if reqErr, ok := err.(awserr.RequestFailure); ok {
			if reqErr.StatusCode() == 404 {
				log.Printf("File '%s' does not exist in bucket '%s'\n", dest, client.s3cliConfig.BucketName)
				return false, nil
			}
		}
		return false, err
	}

	log.Printf("File '%s' exist in bucket '%s'\n", dest, client.s3cliConfig.BucketName)
	return true, nil
}
