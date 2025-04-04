package integration

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/cloudfoundry/bosh-s3cli/client"
	"github.com/cloudfoundry/bosh-s3cli/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	. "github.com/onsi/gomega" //nolint:staticcheck
)

// AssertLifecycleWorks tests the main blobstore object lifecycle from creation to deletion
func AssertLifecycleWorks(s3CLIPath string, cfg *config.S3Cli) {
	expectedString := GenerateRandomString()
	s3Filename := GenerateRandomString()

	configPath := MakeConfigFile(cfg)
	defer os.Remove(configPath) //nolint:errcheck

	contentFile := MakeContentFile(expectedString)
	defer os.Remove(contentFile) //nolint:errcheck

	s3CLISession, err := RunS3CLI(s3CLIPath, configPath, "put", contentFile, s3Filename)
	Expect(err).ToNot(HaveOccurred())
	Expect(s3CLISession.ExitCode()).To(BeZero())

	if len(cfg.FolderName) != 0 {
		folderName := cfg.FolderName
		cfg.FolderName = ""
		noFolderConfigPath := MakeConfigFile(cfg)
		defer os.Remove(noFolderConfigPath) //nolint:errcheck

		s3CLISession, err :=
			RunS3CLI(s3CLIPath, noFolderConfigPath, "exists", fmt.Sprintf("%s/%s", folderName, s3Filename))
		Expect(err).ToNot(HaveOccurred())
		Expect(s3CLISession.ExitCode()).To(BeZero())
	}

	s3CLISession, err = RunS3CLI(s3CLIPath, configPath, "exists", s3Filename)
	Expect(err).ToNot(HaveOccurred())
	Expect(s3CLISession.ExitCode()).To(BeZero())
	Expect(s3CLISession.Err.Contents()).To(MatchRegexp("File '.*' exists in bucket '.*'"))

	tmpLocalFile, err := os.CreateTemp("", "s3cli-download")
	Expect(err).ToNot(HaveOccurred())
	err = tmpLocalFile.Close()
	Expect(err).ToNot(HaveOccurred())
	defer os.Remove(tmpLocalFile.Name()) //nolint:errcheck

	s3CLISession, err = RunS3CLI(s3CLIPath, configPath, "get", s3Filename, tmpLocalFile.Name())
	Expect(err).ToNot(HaveOccurred())
	Expect(s3CLISession.ExitCode()).To(BeZero())

	gottenBytes, err := os.ReadFile(tmpLocalFile.Name())
	Expect(err).ToNot(HaveOccurred())
	Expect(string(gottenBytes)).To(Equal(expectedString))

	s3CLISession, err = RunS3CLI(s3CLIPath, configPath, "delete", s3Filename)
	Expect(err).ToNot(HaveOccurred())
	Expect(s3CLISession.ExitCode()).To(BeZero())

	s3CLISession, err = RunS3CLI(s3CLIPath, configPath, "exists", s3Filename)
	Expect(err).ToNot(HaveOccurred())
	Expect(s3CLISession.ExitCode()).To(Equal(3))
	Expect(s3CLISession.Err.Contents()).To(MatchRegexp("File '.*' does not exist in bucket '.*'"))
}

func AssertOnPutFailures(s3CLIPath string, cfg *config.S3Cli, content, errorMessage string) {
	s3Filename := GenerateRandomString()
	sourceContent := strings.NewReader(content)

	configPath := MakeConfigFile(cfg)
	defer os.Remove(configPath) //nolint:errcheck

	configFile, err := os.Open(configPath)
	Expect(err).ToNot(HaveOccurred())

	s3Config, err := config.NewFromReader(configFile)
	Expect(err).ToNot(HaveOccurred())

	s3Client, err := client.NewAwsS3Client(&s3Config)
	if err != nil {
		log.Fatalln(err)
	}

	// Here we cause every other "part" to fail by way of SHA mismatch,
	// thereby triggering the MultiUploadFailure which causes retries.
	part := 0
	s3Client.Handlers.Build.PushBack(func(r *request.Request) {
		if r.Operation.Name == "UploadPart" {
			if part%2 == 0 {
				r.HTTPRequest.Header.Set("X-Amz-Content-Sha256", "000")
			}
			part++
		}
	})

	blobstoreClient := client.New(s3Client, &s3Config)

	err = blobstoreClient.Put(sourceContent, s3Filename)
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring(errorMessage))
}

// AssertPutOptionsApplied asserts that `s3cli put` uploads files with the requested encryption options
func AssertPutOptionsApplied(s3CLIPath string, cfg *config.S3Cli) {
	expectedString := GenerateRandomString()
	s3Filename := GenerateRandomString()

	configPath := MakeConfigFile(cfg)
	defer os.Remove(configPath) //nolint:errcheck

	contentFile := MakeContentFile(expectedString)
	defer os.Remove(contentFile) //nolint:errcheck

	configFile, err := os.Open(configPath)
	Expect(err).ToNot(HaveOccurred())

	s3CLISession, err := RunS3CLI(s3CLIPath, configPath, "put", contentFile, s3Filename) //nolint:ineffassign,staticcheck

	s3Config, err := config.NewFromReader(configFile)
	Expect(err).ToNot(HaveOccurred())

	s3Client, _ := client.NewAwsS3Client(&s3Config) //nolint:errcheck
	resp, err := s3Client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(cfg.BucketName),
		Key:    aws.String(s3Filename),
	})
	Expect(err).ToNot(HaveOccurred())
	Expect(s3CLISession.ExitCode()).To(BeZero())

	if cfg.ServerSideEncryption == "" {
		Expect(resp.ServerSideEncryption).To(Or(BeNil(), HaveValue(Equal("AES256"))))
	} else {
		Expect(*resp.ServerSideEncryption).To(Equal(cfg.ServerSideEncryption))
	}
}

// AssertGetNonexistentFails asserts that `s3cli get` on a non-existent object will fail
func AssertGetNonexistentFails(s3CLIPath string, cfg *config.S3Cli) {
	configPath := MakeConfigFile(cfg)
	defer os.Remove(configPath) //nolint:errcheck

	s3CLISession, err := RunS3CLI(s3CLIPath, configPath, "get", "non-existent-file", "/dev/null")
	Expect(err).ToNot(HaveOccurred())
	Expect(s3CLISession.ExitCode()).ToNot(BeZero())
	Expect(s3CLISession.Err.Contents()).To(ContainSubstring("NoSuchKey"))
}

// AssertDeleteNonexistentWorks asserts that `s3cli delete` on a non-existent
// object exits with status 0 (tests idempotency)
func AssertDeleteNonexistentWorks(s3CLIPath string, cfg *config.S3Cli) {
	configPath := MakeConfigFile(cfg)
	defer os.Remove(configPath) //nolint:errcheck

	s3CLISession, err := RunS3CLI(s3CLIPath, configPath, "delete", "non-existent-file")
	Expect(err).ToNot(HaveOccurred())
	Expect(s3CLISession.ExitCode()).To(BeZero())
}

func AssertOnMultipartUploads(s3CLIPath string, cfg *config.S3Cli, content string) {
	s3Filename := GenerateRandomString()
	sourceContent := strings.NewReader(content)

	configPath := MakeConfigFile(cfg)
	defer os.Remove(configPath) //nolint:errcheck

	configFile, err := os.Open(configPath)
	Expect(err).ToNot(HaveOccurred())

	s3Config, err := config.NewFromReader(configFile)
	Expect(err).ToNot(HaveOccurred())

	s3Client, err := client.NewAwsS3Client(&s3Config)
	if err != nil {
		log.Fatalln(err)
	}

	tracedS3, calls := traceS3(s3Client)

	blobstoreClient := client.New(tracedS3, &s3Config)

	err = blobstoreClient.Put(sourceContent, s3Filename)
	Expect(err).ToNot(HaveOccurred())

	switch cfg.Host {
	case "storage.googleapis.com":
		Expect(*calls).To(Equal([]string{"PutObject"}))
	default:
		Expect(*calls).To(Equal([]string{"CreateMultipart", "UploadPart", "UploadPart", "CompleteMultipart"}))
	}
}

// AssertOnSignedURLs asserts on using signed URLs for upload and download
func AssertOnSignedURLs(s3CLIPath string, cfg *config.S3Cli) {
	s3Filename := GenerateRandomString()

	configPath := MakeConfigFile(cfg)
	defer os.Remove(configPath) //nolint:errcheck

	configFile, err := os.Open(configPath)
	Expect(err).ToNot(HaveOccurred())

	s3Config, err := config.NewFromReader(configFile)
	Expect(err).ToNot(HaveOccurred())

	s3Client, err := client.NewAwsS3Client(&s3Config)
	if err != nil {
		log.Fatalln(err)
	}

	tracedS3, _ := traceS3(s3Client)

	blobstoreClient := client.New(tracedS3, &s3Config)

	regex := `(?m)((([A-Za-z]{3,9}:(?:\/\/?)?)(?:[-;:&=\+\$,\w]+@)?[A-Za-z0-9.-]+(:[0-9]+)?|(?:www.|[-;:&=\+\$,\w]+@)[A-Za-z0-9.-]+)((?:\/[\+~%\/.\w-_]*)?\??(?:[-\+=&;%@.\w_]*)#?(?:[\w]*))?)`

	// get
	url, err := blobstoreClient.Sign(s3Filename, "get", 1*time.Minute)
	Expect(err).ToNot(HaveOccurred())
	Expect(url).To(MatchRegexp(regex))

	// put
	url, err = blobstoreClient.Sign(s3Filename, "put", 1*time.Minute)
	Expect(err).ToNot(HaveOccurred())
	Expect(url).To(MatchRegexp(regex))
}

func traceS3(svc *s3.S3) (*s3.S3, *[]string) {
	var m sync.Mutex
	calls := []string{}
	partNum := 0

	svc.Handlers.Send.Clear()
	svc.Handlers.Send.PushBack(func(r *request.Request) {
		m.Lock()
		defer m.Unlock()

		r.HTTPResponse = &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader([]byte{})),
		}

		switch data := r.Data.(type) {
		case *s3.CreateMultipartUploadOutput:
			calls = append(calls, "CreateMultipart")
			data.UploadId = aws.String("UPLOAD-ID")
		case *s3.UploadPartOutput:
			calls = append(calls, "UploadPart")
			partNum++
			data.ETag = aws.String(fmt.Sprintf("ETAG%d", partNum))
		case *s3.CompleteMultipartUploadOutput:
			calls = append(calls, "CompleteMultipart")
			data.Location = aws.String("https://location")
			data.VersionId = aws.String("VERSION-ID")
		case *s3.PutObjectOutput:
			calls = append(calls, "PutObject")
			data.VersionId = aws.String("VERSION-ID")
		}
	})

	return svc, &calls
}
