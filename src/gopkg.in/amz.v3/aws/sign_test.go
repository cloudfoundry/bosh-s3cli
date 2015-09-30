package aws

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"time"

	. "gopkg.in/check.v1"
)

var _ = Suite(&SigningSuite{})

type SigningSuite struct{}

// TODO(katco-): The signing methodology is a "one size fits all"
// approach. The hashes we check against don't include headers that
// are added in as requisite parts for S3. That doesn't mean the tests
// are invalid, or that signing is broken for these examples, but as
// long as we're adding heads in, it's impossible to know what the new
// signature should be. We should revaluate these later. See:
// https://github.com/go-amz/amz/issues/29
const v4skipReason = `Extra headers present - cannot predict generated signature (issue #29).`

// EC2 ReST authentication docs: http://goo.gl/fQmAN
var testAuth = Auth{"user", "secret"}

func (s *SigningSuite) TestV4SignedUrl(c *C) {

	auth := Auth{"AKIAIOSFODNN7EXAMPLE", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"}
	req, err := http.NewRequest("GET", "https://examplebucket.s3.amazonaws.com/test.txt", nil)
	req.Header.Add("date", "Fri, 24 May 2013 00:00:00 GMT")
	c.Assert(err, IsNil)
	err = SignV4URL(req, auth, USEast.Name, "s3", 86400*time.Second)
	c.Assert(err, IsNil)

	c.Check(req.URL.String(), Equals, "https://examplebucket.s3.amazonaws.com/test.txt?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKIAIOSFODNN7EXAMPLE%2F20130524%2Fus-east-1%2Fs3%2Faws4_request&X-Amz-Date=20130524T000000Z&X-Amz-Expires=86400&X-Amz-Signature=aeeed9bbccd4d02ee5c0109b86d86835f995330da4c265957d157751f604d404&X-Amz-SignedHeaders=host")
}

func (s *SigningSuite) TestV4SignedUrlReserved(c *C) {

	auth := Auth{"AKIAIOSFODNN7EXAMPLE", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"}
	req, err := http.NewRequest("GET", "https://examplebucket.s3.amazonaws.com/some:reserved,characters", nil)
	req.Header.Add("date", "Fri, 24 May 2013 00:00:00 GMT")
	c.Assert(err, IsNil)
	err = SignV4URL(req, auth, USEast.Name, "s3", 86400*time.Second)
	c.Assert(err, IsNil)

	c.Check(req.URL.String(), Equals, "https://examplebucket.s3.amazonaws.com/some:reserved,characters?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKIAIOSFODNN7EXAMPLE%2F20130524%2Fus-east-1%2Fs3%2Faws4_request&X-Amz-Date=20130524T000000Z&X-Amz-Expires=86400&X-Amz-Signature=ac81e03593d6fc22ac045b9353b0242da755be2af80b981eb13917d8b9cf20a4&X-Amz-SignedHeaders=host")
}

func (s *SigningSuite) TestV4StringToSign(c *C) {

	mockTime, err := time.Parse(time.RFC3339, "2011-09-09T23:36:00Z")
	c.Assert(err, IsNil)
	stringToSign, err := stringToSign(
		mockTime,
		"3511de7e95d28ecd39e9513b642aee07e54f4941150d8df8bf94b328ef7e55e2",
		"20110909/us-east-1/iam/aws4_request",
	)
	c.Assert(err, IsNil)

	const expected = `AWS4-HMAC-SHA256
20110909T233600Z
20110909/us-east-1/iam/aws4_request
3511de7e95d28ecd39e9513b642aee07e54f4941150d8df8bf94b328ef7e55e2`
	c.Assert(stringToSign, Equals, expected)
}

func (s *SigningSuite) TestV4CanonicalRequest(c *C) {

	c.Skip(v4skipReason)

	body := new(bytes.Buffer)
	_, err := fmt.Fprint(body, "Action=ListUsers&Version=2010-05-08")
	c.Assert(err, IsNil)

	req, err := http.NewRequest("POST", "https://iam.amazonaws.com", body)
	c.Assert(err, IsNil)

	req.Header.Add("content-type", "application/x-www-form-urlencoded; charset=utf-8")
	req.Header.Add("host", req.URL.Host)
	req.Header.Add("x-amz-date", "20110909T233600Z")

	canonReq, canonReqHash, _, err := canonicalRequest(
		req,
		sha256Hasher,
		true,
	)
	c.Assert(err, IsNil)

	const expected = `POST
/

content-type:application/x-www-form-urlencoded; charset=utf-8
host:iam.amazonaws.com
x-amz-date:20110909T233600Z

content-type;host;x-amz-date
b6359072c78d70ebee1e81adcbab4f01bf2c23245fa365ef83fe8f1f955085e2`

	c.Assert(canonReq, Equals, expected)
	c.Assert(canonReqHash, Equals, "3511de7e95d28ecd39e9513b642aee07e54f4941150d8df8bf94b328ef7e55e2")
}

func (s *SigningSuite) TestV4SigningKey(c *C) {

	c.Skip(v4skipReason)

	mockTime, err := time.Parse(time.RFC3339, "2011-09-09T23:36:00Z")
	c.Assert(err, IsNil)
	c.Assert(
		fmt.Sprintf("%v", signingKey(mockTime, testAuth.SecretKey, USEast.Name, "iam")),
		Equals,
		"[152 241 216 137 254 196 244 66 26 220 82 43 171 12 225 248 46 105 41 194 98 237 21 229 169 76 144 239 209 227 176 231]")
}

func (s *SigningSuite) TestV4BasicSignatureV4(c *C) {

	c.Skip(v4skipReason)

	body := new(bytes.Buffer)

	req, err := http.NewRequest("POST / http/1.1", "https://host.foo.com", body)
	c.Assert(err, IsNil)

	req.Header.Add("Host", req.URL.Host)
	req.Header.Add("Date", "Mon, 09 Sep 2011 23:36:00 GMT")

	testAuth := Auth{
		AccessKey: "AKIDEXAMPLE",
		SecretKey: "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY",
	}
	err = SignV4(req, testAuth, USEast.Name, "host")
	c.Assert(err, IsNil)

	c.Assert(req.Header.Get("Authorization"), Equals, `AWS4-HMAC-SHA256 Credential=AKIDEXAMPLE/20110909/us-east-1/host/aws4_request,SignedHeaders=date;host,Signature=22902d79e148b64e7571c3565769328423fe276eae4b26f83afceda9e767f726`)
}

func (s *SigningSuite) TestV4MoreCompleteSignature(c *C) {

	req, err := http.NewRequest("GET", "https://examplebucket.s3.amazonaws.com/test.txt", nil)
	c.Assert(err, IsNil)

	req.Header.Set("x-amz-date", "20130524T000000Z")
	req.Header.Set("Range", "bytes=0-9")

	testAuth := Auth{
		AccessKey: "AKIAIOSFODNN7EXAMPLE",
		SecretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
	}
	err = SignV4(req, testAuth, USEast.Name, "s3")
	c.Assert(err, IsNil)
	c.Check(req.Header.Get("Authorization"), Equals, "AWS4-HMAC-SHA256 Credential=AKIAIOSFODNN7EXAMPLE/20130524/us-east-1/s3/aws4_request, SignedHeaders=host;range;x-amz-content-sha256;x-amz-date, Signature=f0e8bdb87c964420e857bd35b5d6ed310bd44f0170aba48dd91039c6036bdb41")
}

//
// v2 Tests
//

func (s *SigningSuite) TestV2BasicSignature(c *C) {
	req, err := http.NewRequest("GET", "http://localhost/path", nil)
	c.Assert(err, IsNil)

	SignV2(req, testAuth)

	query := req.URL.Query()

	c.Assert(query.Get("SignatureVersion"), Equals, "2")
	c.Assert(query.Get("SignatureMethod"), Equals, "HmacSHA256")
	expected := "6lSe5QyXum0jMVc7cOUz32/52ZnL7N5RyKRk/09yiK4="
	c.Assert(query.Get("Signature"), Equals, expected)
}

func (s *SigningSuite) TestV2ParamSignature(c *C) {

	req, err := http.NewRequest("GET", "http://localhost/path", nil)
	c.Assert(err, IsNil)

	query := req.URL.Query()
	for i := 1; i <= 3; i++ {
		query.Add(fmt.Sprintf("param%d", i), fmt.Sprintf("value%d", i))
	}
	req.URL.RawQuery = query.Encode()

	SignV2(req, testAuth)

	expected := "XWOR4+0lmK8bD8CGDGZ4kfuSPbb2JibLJiCl/OPu1oU="
	c.Assert(req.URL.Query().Get("Signature"), Equals, expected)
}

func (s *SigningSuite) TestV2ManyParams(c *C) {

	req, err := http.NewRequest("GET", "http://localhost/path", nil)
	c.Assert(err, IsNil)

	query := req.URL.Query()
	orderedVals := []int{10, 2, 3, 4, 5, 6, 7, 8, 9, 1}
	for i, val := range orderedVals {
		query.Add(fmt.Sprintf("param%d", i+1), fmt.Sprintf("value%d", val))
	}
	req.URL.RawQuery = query.Encode()

	SignV2(req, testAuth)

	expected := "di0sjxIvezUgQ1SIL6i+C/H8lL+U0CQ9frLIak8jkVg="
	c.Assert(req.URL.Query().Get("Signature"), Equals, expected)
}

func (s *SigningSuite) TestV2Escaping(c *C) {

	req, err := http.NewRequest("GET", "http://localhost/path", nil)
	c.Assert(err, IsNil)

	query := req.URL.Query()
	query.Add("Nonce", "+ +")
	req.URL.RawQuery = query.Encode()

	err = SignV2(req, testAuth)
	c.Assert(err, IsNil)

	query = req.URL.Query()
	c.Assert(query.Get("Nonce"), Equals, "+ +")

	expected := "bqffDELReIqwjg/W0DnsnVUmfLK4wXVLO4/LuG+1VFA="
	c.Assert(query.Get("Signature"), Equals, expected)
}

func (s *SigningSuite) TestV2SignatureExample1(c *C) {

	req, err := http.NewRequest("GET", "http://sdb.amazonaws.com/", nil)
	c.Assert(err, IsNil)

	query := req.URL.Query()
	query.Add("Timestamp", "2009-02-01T12:53:20+00:00")
	query.Add("Version", "2007-11-07")
	query.Add("Action", "ListDomains")
	req.URL.RawQuery = query.Encode()

	SignV2(req, Auth{"access", "secret"})

	expected := "okj96/5ucWBSc1uR2zXVfm6mDHtgfNv657rRtt/aunQ="
	c.Assert(req.URL.Query().Get("Signature"), Equals, expected)
}

// Tests example from:
// http://docs.aws.amazon.com/general/latest/gr/signature-version-2.html
// Specifically, good for testing case when URL does not contain a /
func (s *SigningSuite) TestV2SignatureTutorialExample(c *C) {

	req, err := http.NewRequest("GET", "https://elasticmapreduce.amazonaws.com/", nil)
	c.Assert(err, IsNil)

	query := req.URL.Query()
	query.Add("Timestamp", "2011-10-03T15:19:30")
	query.Add("Version", "2009-03-31")
	query.Add("Action", "DescribeJobFlows")
	req.URL.RawQuery = query.Encode()

	testAuth := Auth{"AKIAIOSFODNN7EXAMPLE", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"}
	err = SignV2(req, testAuth)
	c.Assert(err, IsNil)
	c.Assert(req.URL.Query().Get("Signature"), Equals, "i91nKc4PWAt0JJIdXwz9HxZCJDdiy6cf/Mj6vPxyYIs=")
}

// https://bugs.launchpad.net/goamz/+bug/1022749
func (s *SigningSuite) TestSignatureWithEndpointPath(c *C) {

	req, err := http.NewRequest("GET", "http://localhost:4444/services/Cloud", nil)
	c.Assert(err, IsNil)

	queryStr := req.URL.Query()
	queryStr.Add("Action", "RebootInstances")
	queryStr.Add("Version", "2011-12-15")
	queryStr.Add("InstanceId.1", "i-10a64379")
	queryStr.Add("Timestamp", time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC).In(time.UTC).Format(time.RFC3339))
	req.URL.RawQuery = queryStr.Encode()

	err = SignV2(req, Auth{"abc", "123"})
	c.Assert(err, IsNil)
	c.Assert(req.URL.Query().Get("Signature"), Equals, "gdG/vEm+c6ehhhfkrJy3+wuVzw/rzKR42TYelMwti7M=")
	err = req.ParseForm()
	c.Assert(err, IsNil)
	c.Assert(req.Form["Signature"], DeepEquals, []string{"gdG/vEm+c6ehhhfkrJy3+wuVzw/rzKR42TYelMwti7M="})
}

func (s *SigningSuite) TestSignS3ObjectGet(c *C) {

	req, err := http.NewRequest("GET", "https://johnsmith.s3.amazonaws.com/photos/puppy.jpg", nil)
	c.Assert(err, IsNil)

	req.Header = map[string][]string{
		"Host": {"johnsmith.s3.amazonaws.com"},
		"Date": {"Tue, 27 Mar 2007 19:36:42 +0000"},
	}
	SignS3(req, testAuth)
	expected := "AWS user:cdkJxumqt+8khyQTkjvGxh/br+g="
	c.Assert(req.Header["Authorization"], DeepEquals, []string{expected})
}

func (s *SigningSuite) TestSignS3ObjectPut(c *C) {

	req, err := http.NewRequest("PUT", "https://johnsmith.s3.amazonaws.com/photos/puppy.jpg", nil)
	c.Assert(err, IsNil)

	req.Header = map[string][]string{
		"Host":           {"johnsmith.s3.amazonaws.com"},
		"Date":           {"Tue, 27 Mar 2007 21:15:45 +0000"},
		"Content-Type":   {"image/jpeg"},
		"Content-Length": {"94328"},
	}
	SignS3(req, testAuth)
	expected := "AWS user:NOBBgFhqYI0D03soKSRLjpqglqU="
	c.Assert(req.Header["Authorization"], DeepEquals, []string{expected})
}

func (s *SigningSuite) TestSignS3List(c *C) {
	req, err := http.NewRequest("GET", "https://johnsmith.s3.amazonaws.com/", nil)
	c.Assert(err, IsNil)

	params := url.Values{
		"prefix":   {"photos"},
		"max-keys": {"50"},
		"marker":   {"puppy"},
	}
	req.URL.RawQuery = params.Encode()
	req.Header = map[string][]string{
		"Host":       {"johnsmith.s3.amazonaws.com"},
		"Date":       {"Tue, 27 Mar 2007 19:42:41 +0000"},
		"User-Agent": {"Mozilla/5.0"},
	}
	SignS3(req, testAuth)
	expected := "AWS user:zF/SIE9rxTVUDrDaDE+yMfHYX40="
	c.Assert(req.Header["Authorization"], DeepEquals, []string{expected})
}

func (s *SigningSuite) TestSignS3Fetch(c *C) {
	req, err := http.NewRequest("GET", "https://johnsmith.s3.amazonaws.com/", nil)
	c.Assert(err, IsNil)

	params := url.Values{
		"acl": {""},
	}
	req.URL.RawQuery = params.Encode()
	req.Header = map[string][]string{
		"Host": {"johnsmith.s3.amazonaws.com"},
		"Date": {"Tue, 27 Mar 2007 19:44:46 +0000"},
	}
	SignS3(req, testAuth)
	expected := "AWS user:8DUW4zS1ew5V7bf68HbV4GCh4eM="
	c.Assert(req.Header["Authorization"], DeepEquals, []string{expected})
}

func (s *SigningSuite) TestSignS3Delete(c *C) {
	req, err := http.NewRequest("DELETE", "https://johnsmith.s3.amazonaws.com/photos/puppy.jpg", nil)
	c.Assert(err, IsNil)

	req.Header = map[string][]string{
		"Host":       {"s3.amazonaws.com"},
		"Date":       {"Tue, 27 Mar 2007 21:20:27 +0000"},
		"User-Agent": {"dotnet"},
		"x-amz-date": {"Tue, 27 Mar 2007 21:20:26 +0000"},
	}
	SignS3(req, testAuth)
	expected := "AWS user:qAH1zJTA1GQvMxmeLnQj2aa6MuM="
	c.Assert(req.Header["Authorization"], DeepEquals, []string{expected})
}

func (s *SigningSuite) TestSignS3Upload(c *C) {
	req, err := http.NewRequest("PUT", "https://s3.amazonaws.com/static.johnsmith.net/db-backup.dat.gz", nil)
	c.Assert(err, IsNil)

	req.Header = map[string][]string{
		"Host":                         {"static.johnsmith.net:8080"},
		"Date":                         {"Tue, 27 Mar 2007 21:06:08 +0000"},
		"User-Agent":                   {"curl/7.15.5"},
		"x-amz-acl":                    {"public-read"},
		"content-type":                 {"application/x-download"},
		"Content-MD5":                  {"4gJE4saaMU4BqNR0kLY+lw=="},
		"X-Amz-Meta-ReviewedBy":        {"joe@johnsmith.net,jane@johnsmith.net"},
		"X-Amz-Meta-FileChecksum":      {"0x02661779"},
		"X-Amz-Meta-ChecksumAlgorithm": {"crc32"},
		"Content-Disposition":          {"attachment; filename=database.dat"},
		"Content-Encoding":             {"gzip"},
		"Content-Length":               {"5913339"},
	}
	SignS3(req, testAuth)
	expected := "AWS user:WKK7r4VCEJEF/vYlQ5x8r007h9U="
	c.Assert(req.Header["Authorization"], DeepEquals, []string{expected})
}

func (s *SigningSuite) TestSignS3ListAllMyBuckets(c *C) {
	req, err := http.NewRequest("GET", "https://s3.amazonaws.com/", nil)
	c.Assert(err, IsNil)

	req.Header = map[string][]string{
		"Host": {"s3.amazonaws.com"},
		"Date": {"Wed, 28 Mar 2007 01:29:59 +0000"},
	}
	SignS3(req, testAuth)
	expected := "AWS user:bGEYF4XnLpK/yepipZL5hTF2sM8="
	c.Assert(req.Header["Authorization"], DeepEquals, []string{expected})
}

func (s *SigningSuite) TestSignS3UnicodeKeys(c *C) {
	req, err := http.NewRequest("GET", "https://dictionary.s3.amazonaws.com/fran%C3%A7ais/pr%c3%a9f%c3%a8re", nil)
	c.Assert(err, IsNil)

	req.Header = map[string][]string{
		"Host": {"s3.amazonaws.com"},
		"Date": {"Wed, 28 Mar 2007 01:49:49 +0000"},
	}
	SignS3(req, testAuth)
	expected := "AWS user:lNzA+bvy4cUBPqbt5uOvJwMlpEw="
	c.Assert(req.Header["Authorization"], DeepEquals, []string{expected})
}
