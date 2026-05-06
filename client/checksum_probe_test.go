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

func (s *fakeS3Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	headers := map[string][]string{}
	for _, key := range []string{
		"Authorization",
		"Content-Encoding",
		"Content-Length",
		"X-Amz-Checksum-Algorithm",
		"X-Amz-Content-Sha256",
		"X-Amz-Decoded-Content-Length",
		"X-Amz-Sdk-Checksum-Algorithm",
		"X-Amz-Trailer",
		"X-Amz-Checksum-Crc32",
	} {
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

func TestChecksumProbe(t *testing.T) {
	t.Parallel()

	type scenario struct {
		Name      string
		Overrides map[string]any
		Size      int
	}

	const mib = 1024 * 1024
	scenarios := []scenario{
		{
			Name: "single default checksum options",
			Size: 128 * 1024,
		},
		{
			Name: "single all checksum options disabled",
			Overrides: map[string]any{
				"request_checksum_calculation_enabled":          false,
				"response_checksum_calculation_enabled":         false,
				"uploader_request_checksum_calculation_enabled": false,
			},
			Size: 128 * 1024,
		},
		{
			Name: "multipart default checksum options",
			Size: 6 * mib,
		},
		{
			Name: "multipart client checksums disabled but uploader still enabled",
			Overrides: map[string]any{
				"request_checksum_calculation_enabled":  false,
				"response_checksum_calculation_enabled": false,
			},
			Size: 6 * mib,
		},
		{
			Name: "multipart all checksum options disabled",
			Overrides: map[string]any{
				"request_checksum_calculation_enabled":          false,
				"response_checksum_calculation_enabled":         false,
				"uploader_request_checksum_calculation_enabled": false,
			},
			Size: 6 * mib,
		},
	}

	for _, sc := range scenarios {
		sc := sc
		t.Run(sc.Name, func(t *testing.T) {
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
			for key, value := range sc.Overrides {
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
			payload := bytes.NewReader(bytes.Repeat([]byte("a"), sc.Size))
			if err := blobstoreClient.Put(payload, "blob"); err != nil {
				t.Fatal(err)
			}

			assertChecksumProbe(t, sc.Name, fake.snapshot())

			out, err := json.MarshalIndent(fake.snapshot(), "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("\n%s", out)
		})
	}
}

func assertChecksumProbe(t *testing.T, scenario string, requests []capturedS3Request) {
	t.Helper()

	switch scenario {
	case "single default checksum options":
		if len(requests) != 1 {
			t.Fatalf("expected one request, got %d", len(requests))
		}
		assertHeader(t, requests[0], "Content-Encoding", "aws-chunked")
		assertHeader(t, requests[0], "X-Amz-Content-Sha256", "STREAMING-UNSIGNED-PAYLOAD-TRAILER")
		assertHeader(t, requests[0], "X-Amz-Decoded-Content-Length", "131072")
		assertHeader(t, requests[0], "X-Amz-Trailer", "x-amz-checksum-crc32")

	case "single all checksum options disabled":
		if len(requests) != 1 {
			t.Fatalf("expected one request, got %d", len(requests))
		}
		assertNoHeader(t, requests[0], "Content-Encoding")
		assertNoHeader(t, requests[0], "X-Amz-Trailer")
		assertNoHeader(t, requests[0], "X-Amz-Sdk-Checksum-Algorithm")
		assertNoHeader(t, requests[0], "X-Amz-Checksum-Crc32")
		assertHeader(t, requests[0], "Content-Length", "131072")
		assertHeader(t, requests[0], "X-Amz-Content-Sha256", "UNSIGNED-PAYLOAD")

	case "multipart default checksum options":
		assertMultipartChecksumEnabled(t, requests)

	case "multipart client checksums disabled but uploader still enabled":
		assertMultipartChecksumEnabled(t, requests)

	case "multipart all checksum options disabled":
		assertMultipartChecksumDisabled(t, requests)

	default:
		t.Fatalf("unknown scenario %q", scenario)
	}
}

func assertMultipartChecksumEnabled(t *testing.T, requests []capturedS3Request) {
	t.Helper()

	init := findRequest(t, requests, http.MethodPost, "uploads=")
	assertHeader(t, init, "X-Amz-Checksum-Algorithm", "CRC32")

	parts := findUploadParts(t, requests)
	if len(parts) != 2 {
		t.Fatalf("expected two UploadPart requests, got %d", len(parts))
	}
	for _, part := range parts {
		assertHeader(t, part, "Content-Encoding", "aws-chunked")
		assertHeader(t, part, "X-Amz-Content-Sha256", "STREAMING-UNSIGNED-PAYLOAD-TRAILER")
		assertHeader(t, part, "X-Amz-Sdk-Checksum-Algorithm", "CRC32")
		assertHeader(t, part, "X-Amz-Trailer", "x-amz-checksum-crc32")
	}
}

func assertMultipartChecksumDisabled(t *testing.T, requests []capturedS3Request) {
	t.Helper()

	init := findRequest(t, requests, http.MethodPost, "uploads=")
	assertNoHeader(t, init, "X-Amz-Checksum-Algorithm")

	parts := findUploadParts(t, requests)
	if len(parts) != 2 {
		t.Fatalf("expected two UploadPart requests, got %d", len(parts))
	}
	for _, part := range parts {
		assertNoHeader(t, part, "Content-Encoding")
		assertNoHeader(t, part, "X-Amz-Decoded-Content-Length")
		assertNoHeader(t, part, "X-Amz-Sdk-Checksum-Algorithm")
		assertNoHeader(t, part, "X-Amz-Trailer")
		assertNoHeader(t, part, "X-Amz-Checksum-Crc32")
		assertHeader(t, part, "X-Amz-Content-Sha256", "UNSIGNED-PAYLOAD")
	}
}

func findRequest(t *testing.T, requests []capturedS3Request, method string, queryContains string) capturedS3Request {
	t.Helper()
	for _, request := range requests {
		if request.Method == method && strings.Contains(request.Query, queryContains) {
			return request
		}
	}
	t.Fatalf("missing request method=%s query containing %q", method, queryContains)
	return capturedS3Request{}
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
