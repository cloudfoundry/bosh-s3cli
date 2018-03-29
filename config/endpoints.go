package config

import (
	"regexp"
)

var (
	awsHostRegex      = regexp.MustCompile(`s3[-.](.*)\.amazonaws\.com`)
	alicloudHostRegex = regexp.MustCompile(`^oss-([a-z]+-[a-z]+(-[1-9])?)(-internal)?.aliyuncs.com`)
)

func AWSHostToRegion(host string) string {
	regexMatches := awsHostRegex.FindStringSubmatch(host)

	region := "us-east-1"

	if len(regexMatches) == 2 && regexMatches[1] != "external-1" {
		region = regexMatches[1]
	}

	return region
}

func AlicloudHostToRegion(host string) string {
	regexMatches := alicloudHostRegex.FindStringSubmatch(host)

	if len(regexMatches) == 4 {
		return regexMatches[1]
	}

	return ""
}

var multipartBlacklist = []string{
	"storage.googleapis.com",
}
