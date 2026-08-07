package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/muriiloandrade/finsplitter/pkg/logctx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captureRequestIDHandler records the request ID seen by the downstream handler.
func captureRequestIDHandler(t *testing.T) (http.HandlerFunc, *string) {
	t.Helper()
	var captured string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = logctx.GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	})
	return handler, &captured
}

func TestRequestID_PropagatesProvidedHeader(t *testing.T) {
	handler, captured := captureRequestIDHandler(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health/liveness", nil)
	req.Header.Set(RequestIDHeader, "client-provided-id")

	RequestID(handler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "client-provided-id", *captured)
	assert.Equal(t, "client-provided-id", rec.Header().Get(RequestIDHeader))
}

func TestRequestID_GeneratesUUIDv4WhenHeaderAbsent(t *testing.T) {
	handler, captured := captureRequestIDHandler(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	RequestID(handler).ServeHTTP(rec, req)

	require.NotEmpty(t, *captured, "a request ID must be generated")
	parsed, err := uuid.FromString(*captured)
	require.NoError(t, err, "generated request ID must be a valid UUID")
	assert.Equal(t, byte(4), parsed.Version(), "generated request ID must be a UUID v4")
	assert.Equal(t, *captured, rec.Header().Get(RequestIDHeader))
}

func TestRequestID_DoesNotOverwriteExistingResponseHeader(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set(RequestIDHeader, "downstream-set")
		w.WriteHeader(http.StatusOK)
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	RequestID(handler).ServeHTTP(rec, req)

	// The middleware sets the header before calling next; a downstream
	// handler may overwrite it, which is acceptable behavior.
	assert.Equal(t, "downstream-set", rec.Header().Get(RequestIDHeader))
}
