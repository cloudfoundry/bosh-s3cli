package integration_test

import (
	"os"

	"github.com/cloudfoundry/bosh-s3cli/config"
	"github.com/cloudfoundry/bosh-s3cli/integration"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"fmt"
	"strings"
)

var _ = Describe("Testing only in cn-hangzhou", func() {
	Context("with Alibaba Cloud cn-hangzhou (static creds) configurations", func() {
		accessKeyID := os.Getenv("ACCESS_KEY_ID")
		secretAccessKey := os.Getenv("SECRET_ACCESS_KEY")
		bucketName := os.Getenv("BUCKET_NAME")
		region := os.Getenv("REGION")
		ossHost := fmt.Sprintf("oss-%s.aliyuncs.com", strings.TrimSpace(region))

		BeforeEach(func() {
			Expect(accessKeyID).ToNot(BeEmpty(), "ACCESS_KEY_ID must be set")
			Expect(secretAccessKey).ToNot(BeEmpty(), "SECRET_ACCESS_KEY must be set")
			Expect(bucketName).ToNot(BeEmpty(), "BUCKET_NAME must be set")
			Expect(ossHost).ToNot(BeEmpty(), "OSS_HOST must be set")
		})

		configurations := []TableEntry{
			Entry("with minimal config", &config.S3Cli{
				AccessKeyID:     accessKeyID,
				SecretAccessKey: secretAccessKey,
				BucketName:      bucketName,
				Host:            ossHost,
			}),
			Entry("with region specified", &config.S3Cli{
				AccessKeyID:     accessKeyID,
				SecretAccessKey: secretAccessKey,
				BucketName:      bucketName,
				Host:            ossHost,
				Region:          region,
			}),
			Entry("with port specified", &config.S3Cli{
				AccessKeyID:     accessKeyID,
				SecretAccessKey: secretAccessKey,
				BucketName:      bucketName,
				Host:            ossHost,
				Port:            443,
			}),
		}
		DescribeTable("Blobstore lifecycle works",
			func(cfg *config.S3Cli) { integration.AssertLifecycleWorks(s3CLIPath, cfg) },
			configurations...,
		)
		DescribeTable("Invoking `s3cli get` on a non-existent-key fails",
			func(cfg *config.S3Cli) { integration.AssertGetNonexistentFails(s3CLIPath, cfg) },
			configurations...,
		)
		DescribeTable("Invoking `s3cli delete` on a non-existent-key does not fail",
			func(cfg *config.S3Cli) { integration.AssertDeleteNonexistentWorks(s3CLIPath, cfg) },
			configurations...,
		)
	})
})
