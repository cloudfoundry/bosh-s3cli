package integration

import (
	"context"
	"sync/atomic"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// CreateUploadPartTracker creates an Initialize middleware that tracks upload parts
func createUploadPartTracker() middleware.InitializeMiddleware {
	var partCounter int64

	return middleware.InitializeMiddlewareFunc("UploadPartTracker", func(
		ctx context.Context, in middleware.InitializeInput, next middleware.InitializeHandler,
	) (
		out middleware.InitializeOutput, metadata middleware.Metadata, err error,
	) {
		// Type switch to check if the input is s3.UploadPartInput
		injectFailure := false
		switch in.Parameters.(type) {
		case *s3.UploadPartInput:
			// Increment counter and mark even-numbered parts for failure
			count := atomic.AddInt64(&partCounter, 1)
			if count%2 == 0 {
				injectFailure = true
			}
		}

		// Store the injectFailure flag in the context for this request
		ctx = context.WithValue(ctx, failureInjectionKey{}, injectFailure)

		// Continue to next middleware
		return next.HandleInitialize(ctx, in)
	})
}

// Context key for failure injection flag
type failureInjectionKey struct{}

// CreateSHACorruptionMiddleware creates a Finalize middleware that corrupts headers
func createSHACorruptionMiddleware() middleware.FinalizeMiddleware {
	return middleware.FinalizeMiddlewareFunc("SHACorruptionMiddleware", func(
		ctx context.Context, in middleware.FinalizeInput, next middleware.FinalizeHandler,
	) (middleware.FinalizeOutput, middleware.Metadata, error) {
		// Check if we should inject a failure based on context value
		if inject, ok := ctx.Value(failureInjectionKey{}).(bool); ok && inject {
			if req, ok := in.Request.(*smithyhttp.Request); ok {
				// Corrupt the SHA256 header to cause upload failure
				req.Header.Set("X-Amz-Content-Sha256", "000")
			}
		}
		return next.HandleFinalize(ctx, in)
	})
}

// S3TracingMiddleware captures S3 operation names for testing
type S3TracingMiddleware struct {
	calls *[]string
}

// CreateS3TracingMiddleware creates a middleware that tracks S3 operation calls
func createS3TracingMiddleware(calls *[]string) *S3TracingMiddleware {
	return &S3TracingMiddleware{calls: calls}
}

// ID returns the middleware identifier
func (m *S3TracingMiddleware) ID() string {
	return "S3TracingMiddleware"
}

// HandleInitialize implements the InitializeMiddleware interface
func (m *S3TracingMiddleware) HandleInitialize(
	ctx context.Context, in middleware.InitializeInput, next middleware.InitializeHandler,
) (middleware.InitializeOutput, middleware.Metadata, error) {
	// Extract operation name from the middleware metadata
	if operationName := middleware.GetStackValue(ctx, "operation"); operationName != nil {
		if opName, ok := operationName.(string); ok {
			*m.calls = append(*m.calls, opName)
		}
	}

	// Try to determine operation from the input type as fallback
	switch in.Parameters.(type) {
	case *s3.CreateMultipartUploadInput:
		*m.calls = append(*m.calls, "CreateMultipart")
	case *s3.UploadPartInput:
		*m.calls = append(*m.calls, "UploadPart")
	case *s3.CompleteMultipartUploadInput:
		*m.calls = append(*m.calls, "CompleteMultipart")
	case *s3.PutObjectInput:
		*m.calls = append(*m.calls, "PutObject")
	case *s3.GetObjectInput:
		*m.calls = append(*m.calls, "GetObject")
	case *s3.DeleteObjectInput:
		*m.calls = append(*m.calls, "DeleteObject")
	case *s3.HeadObjectInput:
		*m.calls = append(*m.calls, "HeadObject")
	}

	return next.HandleInitialize(ctx, in)
}
