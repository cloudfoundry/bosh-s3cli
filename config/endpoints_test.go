package config_test

import (
	"github.com/cloudfoundry/bosh-s3cli/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Endpoints", func() {
	DescribeTable("AWSHostToRegion",
		func(host, region string) {
			Expect(config.AWSHostToRegion(host)).To(Equal(region))
		},
		Entry("us-east-1", "this-should-default", "us-east-1"),
		Entry("us-east-1", "s3.amazonaws.com", "us-east-1"),
		Entry("us-east-1", "s3-external-1.amazonaws.com", "us-east-1"),
		Entry("us-east-2", "s3.us-east-2.amazonaws.com", "us-east-2"),
		Entry("us-east-2", "s3-us-east-2.amazonaws.com", "us-east-2"),
		Entry("cn-north-1", "s3.cn-north-1.amazonaws.com.cn", "cn-north-1"),
		Entry("whatever-region", "s3.whatever-region.amazonaws.com", "whatever-region"),
		Entry("some-region", "s3-some-region.amazonaws.com", "some-region"),
	)

	DescribeTable("AlicloudHostToRegion",
		func(host, region string) {
			Expect(config.AlicloudHostToRegion(host)).To(Equal(region))
		},
		Entry("with internal and number", "oss-country-zone-9-internal.aliyuncs.com", "country-zone-9"),
		Entry("with internal and no number", "oss-sichuan-chengdu-internal.aliyuncs.com", "sichuan-chengdu"),
		Entry("without internal and number", "oss-one-two-1.aliyuncs.com", "one-two-1"),
		Entry("without internal and no number", "oss-country-zone.aliyuncs.com", "country-zone"),
		Entry("not alicoud", "s3-us-east-2.amazonaws.com", ""),
	)
})
