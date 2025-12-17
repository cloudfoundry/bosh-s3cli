package config

import (
	"regexp"
)

var (
	providerRegex = map[string]*regexp.Regexp{
		"aws":      regexp.MustCompile(`^$|(https?://)?s3[-.]?(.*)\.amazonaws\.com(\.cn)?$`),
		"alicloud": regexp.MustCompile(`^(https?://)?oss-([a-z]+-[a-z]+(-[1-9])?)(-internal)?.aliyuncs.com$`),
		"google":   regexp.MustCompile(`^(https?://)?storage.googleapis.com$`),
	}
)

func AWSHostToRegion(host string) string {
	regexMatches := providerRegex["aws"].FindStringSubmatch(host)

	region := defaultAWSRegion

	if len(regexMatches) == 4 && regexMatches[2] != "" && regexMatches[2] != "external-1" {
		region = regexMatches[2]
	}

	return region
}

func AlicloudHostToRegion(host string) string {
	regexMatches := providerRegex["alicloud"].FindStringSubmatch(host)
	if len(regexMatches) == 5 {
		return regexMatches[2]
	}

	return ""
}
