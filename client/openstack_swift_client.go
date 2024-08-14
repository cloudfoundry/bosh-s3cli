package client

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cloudfoundry/bosh-s3cli/config"
)

// awsS3Client encapsulates Openstack Swift specific bloblsstore interactions
type openstackSwiftS3Client struct {
	s3cliConfig *config.S3Cli
}

func (c *openstackSwiftS3Client) Sign(objectID string, action string, expiration time.Duration) (string, error) {
	action = strings.ToUpper(action)
	switch action {
	case "GET", "PUT":
		return c.signedURL(action, objectID, expiration)
	default:
		return "", fmt.Errorf("action not implemented: %s", action)
	}
}

func (c *openstackSwiftS3Client) signedURL(action string, objectID string, expiration time.Duration) (string, error) {
	var url string
	path := fmt.Sprintf("/v1/%s/%s/%s", c.s3cliConfig.SwiftAuthAccount, c.s3cliConfig.BucketName, objectID)

	expires := time.Now().Add(expiration).Unix()
	hmacBody := action + "\n" + strconv.FormatInt(expires, 10) + "\n" + path

	if c.s3cliConfig.OpenStackBlobstoreType == "ceph" {
		h_1 := hmac.New(sha1.New, []byte(c.s3cliConfig.SwiftTempURLKey))
		h_1.Write([]byte(hmacBody))
		signature := hex.EncodeToString(h_1.Sum(nil))
		url = fmt.Sprintf("https://%s/swift%s?temp_url_sig=%s&temp_url_expires=%d", c.s3cliConfig.Host, path, signature, expires)
	} else {
		h_256 := hmac.New(sha256.New, []byte(c.s3cliConfig.SwiftTempURLKey))
		h_256.Write([]byte(hmacBody))
		signature := hex.EncodeToString(h_256.Sum(nil))
		url = fmt.Sprintf("https://%s%s?temp_url_sig=%s&temp_url_expires=%d", c.s3cliConfig.Host, path, signature, expires)
	}

	return url, nil
}
