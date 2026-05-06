package client_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/cloudfoundry/bosh-s3cli/client"
	s3cliconfig "github.com/cloudfoundry/bosh-s3cli/config"
)

type capturedS3Request struct {
	Method        string              `json:"method"`
	Path          string              `json:"path"`
	Query         string              `json:"query"`
	Headers       map[string][]string `json:"headers"`
	BodyBytes     int                 `json:"body_bytes"`
	BodyPrefixHex string              `json:"body_prefix_hex"`
}

type fakeS3Server struct {
	mu       sync.Mutex
	requests []capturedS3Request
}

const (
	headerAuthorization             = "Authorization"
	headerContentEncoding           = "Content-Encoding"
	headerContentLength             = "Content-Length"
	headerChecksumAlgorithm         = "X-Amz-Checksum-Algorithm"
	headerContentSHA256             = "X-Amz-Content-Sha256"
	headerDecodedContentLength      = "X-Amz-Decoded-Content-Length"
	headerSDKChecksumAlgorithm      = "X-Amz-Sdk-Checksum-Algorithm"
	headerTrailer                   = "X-Amz-Trailer"
	headerChecksumCRC32             = "X-Amz-Checksum-Crc32"
	sigV4AuthorizationPrefix        = "AWS4-HMAC-SHA256 "
	unsignedPayload                 = "UNSIGNED-PAYLOAD"
	streamingUnsignedPayloadTrailer = "STREAMING-UNSIGNED-PAYLOAD-TRAILER"
	checksumAlgorithmCRC32          = "CRC32"
	checksumTrailerCRC32            = "x-amz-checksum-crc32"
	contentEncodingAWSChunked       = "aws-chunked"
)

var capturedHeaderNames = []string{
	headerAuthorization,
	headerContentEncoding,
	headerContentLength,
	headerChecksumAlgorithm,
	headerContentSHA256,
	headerDecodedContentLength,
	headerSDKChecksumAlgorithm,
	headerTrailer,
	headerChecksumCRC32,
}

func (s *fakeS3Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	headers := map[string][]string{}
	for _, key := range capturedHeaderNames {
		if values, ok := r.Header[key]; ok {
			headers[key] = append([]string(nil), values...)
		}
	}

	prefix := body
	if len(prefix) > 32 {
		prefix = prefix[:32]
	}

	s.mu.Lock()
	s.requests = append(s.requests, capturedS3Request{
		Method:        r.Method,
		Path:          r.URL.Path,
		Query:         r.URL.RawQuery,
		Headers:       headers,
		BodyBytes:     len(body),
		BodyPrefixHex: fmt.Sprintf("%x", prefix),
	})
	s.mu.Unlock()

	q := r.URL.Query()
	w.Header().Set("Content-Type", "application/xml")

	switch {
	case r.Method == http.MethodPost && strings.Contains(r.URL.RawQuery, "uploads"):
		_, _ = io.WriteString(w, `<InitiateMultipartUploadResult><Bucket>bucket</Bucket><Key>blob</Key><UploadId>upload-1</UploadId></InitiateMultipartUploadResult>`)
	case r.Method == http.MethodPut && q.Get("partNumber") != "":
		w.Header().Set("ETag", fmt.Sprintf(`"part-%s"`, q.Get("partNumber")))
	case r.Method == http.MethodPost && q.Get("uploadId") != "":
		_, _ = io.WriteString(w, `<CompleteMultipartUploadResult><Location>fake</Location><Bucket>bucket</Bucket><Key>blob</Key><ETag>"complete"</ETag></CompleteMultipartUploadResult>`)
	case r.Method == http.MethodPut:
		w.Header().Set("ETag", `"single"`)
	default:
		http.Error(w, "unexpected request", http.StatusBadRequest)
	}
}

func (s *fakeS3Server) snapshot() []capturedS3Request {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]capturedS3Request(nil), s.requests...)
}

func TestPutObjectDefaultChecksumsUseAWSChunkedCRC32Trailer(t *testing.T) {
	t.Parallel()

	requests := putToFakeS3(t, 128*1024, nil)
	assertRequestCount(t, requests, 1)

	put := findPutObject(t, requests)
	assertSigV4Authorization(t, put)
	assertHeader(t, put, headerContentEncoding, contentEncodingAWSChunked)
	assertHeader(t, put, headerContentLength, "131117")
	assertHeader(t, put, headerContentSHA256, streamingUnsignedPayloadTrailer)
	assertHeader(t, put, headerDecodedContentLength, "131072")
	assertHeader(t, put, headerTrailer, checksumTrailerCRC32)
	assertBodyBytes(t, put, 131117)
}

func TestPutObjectWithChecksumsDisabledSendsPlainUnsignedPayload(t *testing.T) {
	t.Parallel()

	requests := putToFakeS3(t, 128*1024, allChecksumOptionsDisabled())
	assertRequestCount(t, requests, 1)

	put := findPutObject(t, requests)
	assertSigV4Authorization(t, put)
	assertNoFlexibleChecksumHeaders(t, put)
	assertHeader(t, put, headerContentLength, "131072")
	assertHeader(t, put, headerContentSHA256, unsignedPayload)
	assertBodyBytes(t, put, 131072)
}

func TestMultipartDefaultChecksumsUseCRC32ForCreateAndParts(t *testing.T) {
	t.Parallel()

	requests := putToFakeS3(t, 6*mib, nil)
	assertRequestCount(t, requests, 4)

	create := findCreateMultipartUpload(t, requests)
	assertSigV4Authorization(t, create)
	assertHeader(t, create, headerChecksumAlgorithm, checksumAlgorithmCRC32)
	assertHeader(t, create, headerContentLength, "0")
	assertBodyBytes(t, create, 0)

	parts := findUploadParts(t, requests)
	assertUploadPartCount(t, parts, 2)
	for _, part := range parts {
		assertSigV4Authorization(t, part)
		assertHeader(t, part, headerContentEncoding, contentEncodingAWSChunked)
		assertHeader(t, part, headerContentSHA256, streamingUnsignedPayloadTrailer)
		assertHeader(t, part, headerSDKChecksumAlgorithm, checksumAlgorithmCRC32)
		assertHeader(t, part, headerTrailer, checksumTrailerCRC32)
	}
	assertUploadPartBodyBytes(t, parts, 5_242_926, 1_048_622)

	complete := findCompleteMultipartUpload(t, requests)
	assertSigV4Authorization(t, complete)
	assertNoFlexibleChecksumHeaders(t, complete)
	assertHeader(t, complete, headerContentLength, "235")
	assertBodyBytes(t, complete, 235)
}

func TestMultipartStillUsesChecksumsWhenOnlyClientChecksumOptionsAreDisabled(t *testing.T) {
	t.Parallel()

	requests := putToFakeS3(t, 6*mib, clientChecksumOptionsDisabled())
	assertRequestCount(t, requests, 4)

	create := findCreateMultipartUpload(t, requests)
	assertSigV4Authorization(t, create)
	assertHeader(t, create, headerChecksumAlgorithm, checksumAlgorithmCRC32)
	assertHeader(t, create, headerContentLength, "0")
	assertBodyBytes(t, create, 0)

	parts := findUploadParts(t, requests)
	assertUploadPartCount(t, parts, 2)
	for _, part := range parts {
		assertSigV4Authorization(t, part)
		assertHeader(t, part, headerContentEncoding, contentEncodingAWSChunked)
		assertHeader(t, part, headerContentSHA256, streamingUnsignedPayloadTrailer)
		assertHeader(t, part, headerSDKChecksumAlgorithm, checksumAlgorithmCRC32)
		assertHeader(t, part, headerTrailer, checksumTrailerCRC32)
	}
	assertUploadPartBodyBytes(t, parts, 5_242_926, 1_048_622)

	complete := findCompleteMultipartUpload(t, requests)
	assertSigV4Authorization(t, complete)
	assertNoFlexibleChecksumHeaders(t, complete)
	assertHeader(t, complete, headerContentLength, "235")
	assertBodyBytes(t, complete, 235)
}

func TestMultipartWithAllChecksumOptionsDisabledSendsPlainUploadParts(t *testing.T) {
	t.Parallel()

	requests := putToFakeS3(t, 6*mib, allChecksumOptionsDisabled())
	assertRequestCount(t, requests, 4)

	create := findCreateMultipartUpload(t, requests)
	assertSigV4Authorization(t, create)
	assertNoFlexibleChecksumHeaders(t, create)
	assertHeader(t, create, headerContentLength, "0")
	assertBodyBytes(t, create, 0)

	parts := findUploadParts(t, requests)
	assertUploadPartCount(t, parts, 2)
	for _, part := range parts {
		assertSigV4Authorization(t, part)
		assertNoFlexibleChecksumHeaders(t, part)
		assertHeader(t, part, headerContentSHA256, unsignedPayload)
	}
	assertUploadPartBodyBytes(t, parts, 5_242_880, 1_048_576)

	complete := findCompleteMultipartUpload(t, requests)
	assertSigV4Authorization(t, complete)
	assertNoFlexibleChecksumHeaders(t, complete)
	assertHeader(t, complete, headerContentLength, "235")
	assertBodyBytes(t, complete, 235)
}

const mib = 1024 * 1024

func putToFakeS3(t *testing.T, size int, overrides map[string]any) []capturedS3Request {
	t.Helper()

	fake := &fakeS3Server{}
	server := httptest.NewTLSServer(fake)
	defer server.Close()

	rawConfig := map[string]any{
		"access_key_id":     "id",
		"secret_access_key": "key",
		"bucket_name":       "bucket",
		"host":              server.URL,
		"region":            "us-east-1",
		"ssl_verify_peer":   false,
	}
	for key, value := range overrides {
		rawConfig[key] = value
	}

	configBytes, err := json.Marshal(rawConfig)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := s3cliconfig.NewFromReader(bytes.NewReader(configBytes))
	if err != nil {
		t.Fatal(err)
	}

	s3Client, err := client.NewAwsS3Client(&cfg)
	if err != nil {
		t.Fatal(err)
	}

	blobstoreClient := client.New(s3Client, &cfg)
	payload := bytes.NewReader(bytes.Repeat([]byte("a"), size))
	if err := blobstoreClient.Put(payload, "blob"); err != nil {
		t.Fatal(err)
	}

	requests := fake.snapshot()
	out, err := json.MarshalIndent(requests, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("\n%s", out)

	return requests
}

func allChecksumOptionsDisabled() map[string]any {
	return map[string]any{
		"request_checksum_calculation_enabled":          false,
		"response_checksum_calculation_enabled":         false,
		"uploader_request_checksum_calculation_enabled": false,
	}
}

func clientChecksumOptionsDisabled() map[string]any {
	return map[string]any{
		"request_checksum_calculation_enabled":  false,
		"response_checksum_calculation_enabled": false,
	}
}

func findPutObject(t *testing.T, requests []capturedS3Request) capturedS3Request {
	t.Helper()
	return findOneRequest(t, requests, "PutObject", func(request capturedS3Request) bool {
		return request.Method == http.MethodPut && strings.Contains(request.Query, "x-id=PutObject")
	})
}

func findCreateMultipartUpload(t *testing.T, requests []capturedS3Request) capturedS3Request {
	t.Helper()
	return findOneRequest(t, requests, "CreateMultipartUpload", func(request capturedS3Request) bool {
		return request.Method == http.MethodPost && strings.Contains(request.Query, "uploads=")
	})
}

func findCompleteMultipartUpload(t *testing.T, requests []capturedS3Request) capturedS3Request {
	t.Helper()
	return findOneRequest(t, requests, "CompleteMultipartUpload", func(request capturedS3Request) bool {
		return request.Method == http.MethodPost && strings.Contains(request.Query, "uploadId=upload-1")
	})
}

func findUploadParts(t *testing.T, requests []capturedS3Request) []capturedS3Request {
	t.Helper()

	var parts []capturedS3Request
	for _, request := range requests {
		if request.Method == http.MethodPut && strings.Contains(request.Query, "partNumber=") {
			parts = append(parts, request)
		}
	}
	return parts
}

func findOneRequest(t *testing.T, requests []capturedS3Request, label string, matches func(capturedS3Request) bool) capturedS3Request {
	t.Helper()

	var found []capturedS3Request
	for _, request := range requests {
		if matches(request) {
			found = append(found, request)
		}
	}
	if len(found) != 1 {
		t.Fatalf("%s: expected exactly one matching request, got %d", label, len(found))
	}
	return found[0]
}

func assertRequestCount(t *testing.T, requests []capturedS3Request, want int) {
	t.Helper()
	if len(requests) != want {
		t.Fatalf("expected %d request(s), got %d", want, len(requests))
	}
}

func assertUploadPartCount(t *testing.T, parts []capturedS3Request, want int) {
	t.Helper()
	if len(parts) != want {
		t.Fatalf("expected %d UploadPart request(s), got %d", want, len(parts))
	}
}

func assertUploadPartBodyBytes(t *testing.T, parts []capturedS3Request, expected ...int) {
	t.Helper()

	remaining := append([]int(nil), expected...)
	for _, part := range parts {
		var matched bool
		for i, size := range remaining {
			if part.BodyBytes == size {
				remaining = append(remaining[:i], remaining[i+1:]...)
				matched = true
				break
			}
		}
		if !matched {
			t.Fatalf("UploadPart %s?%s: unexpected body byte count %d, remaining expected values %v",
				part.Method, part.Query, part.BodyBytes, remaining)
		}
	}

	if len(remaining) != 0 {
		t.Fatalf("expected UploadPart body byte count(s) not observed: %v", remaining)
	}
}

func assertNoFlexibleChecksumHeaders(t *testing.T, request capturedS3Request) {
	t.Helper()

	assertNoHeader(t, request, headerContentEncoding)
	assertNoHeader(t, request, headerChecksumAlgorithm)
	assertNoHeader(t, request, headerDecodedContentLength)
	assertNoHeader(t, request, headerSDKChecksumAlgorithm)
	assertNoHeader(t, request, headerTrailer)
	assertNoHeader(t, request, headerChecksumCRC32)
}

func assertBodyBytes(t *testing.T, request capturedS3Request, want int) {
	t.Helper()
	if request.BodyBytes != want {
		t.Fatalf("%s %s?%s: expected %d body bytes, got %d", request.Method, request.Path, request.Query, want, request.BodyBytes)
	}
}

func assertSigV4Authorization(t *testing.T, request capturedS3Request) {
	t.Helper()
	values := request.Headers[headerAuthorization]
	for _, value := range values {
		if strings.HasPrefix(value, sigV4AuthorizationPrefix) {
			return
		}
	}
	t.Fatalf("%s %s?%s: expected SigV4 authorization, got %v", request.Method, request.Path, request.Query, values)
}

func assertHeader(t *testing.T, request capturedS3Request, key string, want string) {
	t.Helper()
	values := request.Headers[key]
	for _, value := range values {
		if value == want {
			return
		}
	}
	t.Fatalf("%s %s?%s: expected header %s=%q, got %v", request.Method, request.Path, request.Query, key, want, values)
}

func assertNoHeader(t *testing.T, request capturedS3Request, key string) {
	t.Helper()
	if values, ok := request.Headers[key]; ok {
		t.Fatalf("%s %s?%s: expected no header %s, got %v", request.Method, request.Path, request.Query, key, values)
	}
}
