package integration

import (
	"encoding/json"
	"math/rand"
	"os"
	"os/exec"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go/middleware"
	"github.com/cloudfoundry/bosh-s3cli/client"
	"github.com/cloudfoundry/bosh-s3cli/config"

	. "github.com/onsi/ginkgo/v2" //nolint:staticcheck
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

const alphaNum = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// GenerateRandomString generates a random string of desired length (default: 25)
func GenerateRandomString(params ...int) string {
	size := 25
	if len(params) == 1 {
		size = params[0]
	}

	randBytes := make([]byte, size)
	for i := range randBytes {
		randBytes[i] = alphaNum[rand.Intn(len(alphaNum))]
	}
	return string(randBytes)
}

// MakeConfigFile creates a config file from a S3Cli config struct
func MakeConfigFile(cfg *config.S3Cli) string {
	cfgBytes, err := json.Marshal(cfg)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	tmpFile, err := os.CreateTemp("", "s3cli-test")
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	_, err = tmpFile.Write(cfgBytes)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	err = tmpFile.Close()
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	return tmpFile.Name()
}

// MakeContentFile creates a temporary file with content to upload to S3
func MakeContentFile(content string) string {
	tmpFile, err := os.CreateTemp("", "s3cli-test-content")
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	_, err = tmpFile.Write([]byte(content))
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	err = tmpFile.Close()
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	return tmpFile.Name()
}

// RunS3CLI runs the s3cli and outputs the session after waiting for it to finish
func RunS3CLI(s3CLIPath string, configPath string, subcommand string, args ...string) (*gexec.Session, error) {
	cmdArgs := []string{
		"-c",
		configPath,
		subcommand,
	}
	cmdArgs = append(cmdArgs, args...)
	command := exec.Command(s3CLIPath, cmdArgs...)
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	if err != nil {
		return nil, err
	}
	session.Wait(1 * time.Minute)
	return session, nil
}

// CreateS3ClientWithFailureInjection creates an S3 client with failure injection middleware
func CreateS3ClientWithFailureInjection(s3Config *config.S3Cli) (*s3.Client, error) {
	var apiOptions []func(stack *middleware.Stack) error
	// Create tracker once so the middleware shares a single counter across requests.
	uploadPartTracker := createUploadPartTracker()
	apiOptions = append(apiOptions, func(stack *middleware.Stack) error {
		// Add initialize middleware to track UploadPart operations
		if err := stack.Initialize.Add(uploadPartTracker, middleware.Before); err != nil {
			return err
		}
		// Add finalize middleware to corrupt headers after signing
		return stack.Finalize.Add(createSHACorruptionMiddleware(), middleware.After)
	})

	return client.NewAwsS3ClientWithApiOptions(s3Config, apiOptions, true)
}

// CreateTracingS3Client creates an S3 client with tracing middleware
func CreateTracingS3Client(s3Config *config.S3Cli, calls *[]string) (*s3.Client, error) {
	var apiOptions []func(stack *middleware.Stack) error
	// Setup middleware fixing request to Google - they expect the 'accept-encoding' header
	// to not be included in the signature of the request.
	apiOptions = append(apiOptions, client.AddFixAcceptEncodingMiddleware)
	// Use the centralized client creation logic with a custom middleware
	apiOptions = append(apiOptions, func(stack *middleware.Stack) error {
		return stack.Initialize.Add(createS3TracingMiddleware(calls), middleware.Before)
	})

	return client.NewAwsS3ClientWithApiOptions(s3Config, apiOptions, false)
}
