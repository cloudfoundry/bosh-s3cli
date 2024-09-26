package client

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cloudfoundry/bosh-s3cli/config"
)

// awsS3Client encapsulates Openstack Ceph specific bloblsstore interactions
type openstackCephS3Client struct {
	s3cliConfig *config.S3Cli
}

func (c *openstackCephS3Client) Sign(objectID string, action string, expiration time.Duration) (string, error) {
	action = strings.ToUpper(action)
	switch action {
	case "GET", "PUT":
		return c.signedURL(action, objectID, expiration)
	default:
		return "", fmt.Errorf("action not implemented: %s", action)
	}
}

func (c *openstackCephS3Client) signedURL(action string, objectID string, expiration time.Duration) (string, error) {
	path := fmt.Sprintf("/v1/%s/%s/%s", c.s3cliConfig.SwiftAuthAccount, c.s3cliConfig.BucketName, objectID)

	expires := time.Now().Add(expiration).Unix()
	hmacBody := action + "\n" + strconv.FormatInt(expires, 10) + "\n" + path

	h := hmac.New(sha1.New, []byte(c.s3cliConfig.SwiftTempURLKey))
	h.Write([]byte(hmacBody))
	signature := hex.EncodeToString(h.Sum(nil))

	url := fmt.Sprintf("https://%s/swift%s?temp_url_sig=%s&temp_url_expires=%d", c.s3cliConfig.Host, path, signature, expires)

	return url, nil
}
