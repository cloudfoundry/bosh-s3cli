package integration

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cloudfoundry/bosh-s3cli/client"
	"github.com/cloudfoundry/bosh-s3cli/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
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

	s3Client, err := CreateS3ClientWithFailureInjection(&s3Config)
	if err != nil {
		log.Fatalln(err)
	}
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
	Expect(err).ToNot(HaveOccurred())
	Expect(s3CLISession.ExitCode()).To(BeZero())

	s3Config, err := config.NewFromReader(configFile)
	Expect(err).ToNot(HaveOccurred())

	s3Client, err := client.NewAwsS3Client(&s3Config)
	Expect(err).ToNot(HaveOccurred())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(cfg.BucketName),
		Key:    aws.String(s3Filename),
	})
	Expect(err).ToNot(HaveOccurred())

	if cfg.ServerSideEncryption == "" {
		Expect(resp.ServerSideEncryption).To(Or(BeNil(), HaveValue(Equal(types.ServerSideEncryptionAes256))))
	} else {
		Expect(string(resp.ServerSideEncryption)).To(Equal(cfg.ServerSideEncryption))
	}

	// Clean up the uploaded file
	_, err = RunS3CLI(s3CLIPath, configPath, "delete", s3Filename)
	Expect(err).ToNot(HaveOccurred())
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

	// GCP return 404 when trying to delete a non-existent file, others return 204
	switch config.Provider(cfg.Host) {
	case "google":
		Expect(s3CLISession.ExitCode()).ToNot(BeZero())
	default:
		Expect(s3CLISession.ExitCode()).To(BeZero())
	}
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

	// Create S3 client with tracing middleware
	calls := []string{}
	s3Client, err := CreateTracingS3Client(&s3Config, &calls)
	if err != nil {
		log.Fatalln(err)
	}

	blobstoreClient := client.New(s3Client, &s3Config)

	err = blobstoreClient.Put(sourceContent, s3Filename)
	Expect(err).ToNot(HaveOccurred())

	switch cfg.Host {
	case "https://storage.googleapis.com":
		Expect(calls).To(Equal([]string{"PutObject"}))
	default:
		Expect(calls).To(Equal([]string{"CreateMultipart", "UploadPart", "UploadPart", "CompleteMultipart"}))
	}

	// Clean up the uploaded file
	_, err = RunS3CLI(s3CLIPath, configPath, "delete", s3Filename)
	Expect(err).ToNot(HaveOccurred())
}

// AssertOnSignedURLs asserts on using signed URLs for upload and download
func AssertOnSignedURLs(s3CLIPath string, cfg *config.S3Cli) {
	s3Filename := GenerateRandomString()
	expectedContent := GenerateRandomString()

	configPath := MakeConfigFile(cfg)
	defer os.Remove(configPath) //nolint:errcheck

	configFile, err := os.Open(configPath)
	Expect(err).ToNot(HaveOccurred())

	s3Config, err := config.NewFromReader(configFile)
	Expect(err).ToNot(HaveOccurred())

	// Create S3 client with tracing middleware (though signing operations don't need tracing for this test)
	calls := []string{}
	s3Client, err := CreateTracingS3Client(&s3Config, &calls)
	if err != nil {
		log.Fatalln(err)
	}

	blobstoreClient := client.New(s3Client, &s3Config)

	// First upload a test file using regular put operation
	contentFile := MakeContentFile(expectedContent)
	defer os.Remove(contentFile) //nolint:errcheck

	s3CLISession, err := RunS3CLI(s3CLIPath, configPath, "put", contentFile, s3Filename)
	Expect(err).ToNot(HaveOccurred())
	Expect(s3CLISession.ExitCode()).To(BeZero())

	regex := `(?m)((([A-Za-z]{3,9}:(?:\/\/?)?)(?:[-;:&=\+\$,\w]+@)?[A-Za-z0-9.-]+(:[0-9]+)?|(?:www.|[-;:&=\+\$,\w]+@)[A-Za-z0-9.-]+)((?:\/[\+~%\/.\w-_]*)?\??(?:[-\+=&;%@.\w_]*)#?(?:[\w]*))?)`

	// Test GET signed URL
	getURL, err := blobstoreClient.Sign(s3Filename, "get", 1*time.Minute)
	Expect(err).ToNot(HaveOccurred())
	Expect(getURL).To(MatchRegexp(regex))

	// Actually try to download from the GET signed URL
	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Get(getURL)
	Expect(err).ToNot(HaveOccurred())

	body, err := io.ReadAll(resp.Body)
	Expect(err).ToNot(HaveOccurred())
	Expect(string(body)).To(Equal(expectedContent))

	// Test PUT signed URL
	putURL, err := blobstoreClient.Sign(s3Filename+"_put_test", "put", 1*time.Minute)
	Expect(err).ToNot(HaveOccurred())
	Expect(putURL).To(MatchRegexp(regex))

	// Actually try to upload to the PUT signed URL
	testUploadContent := "Test upload content via signed URL"
	putReq, err := http.NewRequest("PUT", putURL, strings.NewReader(testUploadContent))
	Expect(err).ToNot(HaveOccurred())

	putReq.Header.Set("Content-Type", "text/plain")
	putResp, err := httpClient.Do(putReq)
	Expect(err).ToNot(HaveOccurred())
	Expect(putResp.StatusCode).To(Equal(200))

	// Clean up the test files
	_, err = RunS3CLI(s3CLIPath, configPath, "delete", s3Filename)
	Expect(err).ToNot(HaveOccurred())

	_, err = RunS3CLI(s3CLIPath, configPath, "delete", s3Filename+"_put_test")
	Expect(err).ToNot(HaveOccurred())
}
