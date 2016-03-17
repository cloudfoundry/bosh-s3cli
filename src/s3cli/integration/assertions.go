package integration

import (
	"io/ioutil"
	"os"
	"s3cli/config"

	"github.com/onsi/gomega"
)

// AssertLifecycleWorks tests the main blobstore object lifecycle from
// creation to deletion
func AssertLifecycleWorks(s3CLIPath string, cfg *config.S3Cli) {
	expectedString, err := GenerateRandomString()
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	s3Filename, err := GenerateRandomString()
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	configPath := MakeConfigFile(cfg)
	defer func() { _ = os.Remove(configPath) }()

	contentFile := MakeContentFile(expectedString)
	defer func() { _ = os.Remove(contentFile) }()

	s3CLISession, err := RunS3CLI(s3CLIPath, configPath, "put", contentFile, s3Filename)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(s3CLISession.ExitCode()).To(gomega.BeZero())

	s3CLISession, err = RunS3CLI(s3CLIPath, configPath, "exists", s3Filename)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(s3CLISession.ExitCode()).To(gomega.BeZero())
	gomega.Expect(s3CLISession.Out.Contents()).To(gomega.MatchRegexp("File '.*' exists in bucket '.*'"))

	tmpLocalFile, err := ioutil.TempFile("", "s3cli-download")
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	err = tmpLocalFile.Close()
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	defer func() { _ = os.Remove(tmpLocalFile.Name()) }()

	s3CLISession, err = RunS3CLI(s3CLIPath, configPath, "get", s3Filename, tmpLocalFile.Name())
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(s3CLISession.ExitCode()).To(gomega.BeZero())

	gottenBytes, err := ioutil.ReadFile(tmpLocalFile.Name())
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(gottenBytes).To(gomega.Equal(expectedString))

	s3CLISession, err = RunS3CLI(s3CLIPath, configPath, "delete", s3Filename)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(s3CLISession.ExitCode()).To(gomega.BeZero())

	s3CLISession, err = RunS3CLI(s3CLIPath, configPath, "exists", s3Filename)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(s3CLISession.ExitCode()).To(gomega.Equal(3))
	gomega.Expect(s3CLISession.Out.Contents()).To(gomega.MatchRegexp("File '.*' does not exist in bucket '.*'"))
}

// AssertGetNonexistentFails asserts that `s3cli get` on a non-existent object
// will fail
func AssertGetNonexistentFails(s3CLIPath string, cfg *config.S3Cli) {
	nonExistentFile := "non-existent-file"

	configPath := MakeConfigFile(cfg)
	defer func() { _ = os.Remove(configPath) }()

	s3CLISession, err := RunS3CLI(s3CLIPath, configPath, "get", nonExistentFile, nonExistentFile)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(s3CLISession.ExitCode()).ToNot(gomega.BeZero())
	gomega.Expect(s3CLISession.Out.Contents()).To(gomega.ContainSubstring("NoSuchKey"))
}

// AssertDeleteNonexistentWorks asserts that `s3cli delete` on a non-existent
// object exits with status 0 (tests idempotency)
func AssertDeleteNonexistentWorks(s3CLIPath string, cfg *config.S3Cli) {
	nonExistentFile := "non-existent-file"

	configPath := MakeConfigFile(cfg)
	defer func() { _ = os.Remove(configPath) }()

	s3CLISession, err := RunS3CLI(s3CLIPath, configPath, "delete", nonExistentFile)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(s3CLISession.ExitCode()).To(gomega.BeZero())
}
