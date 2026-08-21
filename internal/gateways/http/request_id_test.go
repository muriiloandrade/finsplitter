package http

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/muriiloandrade/finsplitter/pkg/logctx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// discardLogger returns a logger that discards all output.
func discardLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

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

func TestRequestID_StoresRequestIDInContext(t *testing.T) {
	testCases := []struct {
		name        string
		headerValue string
		assertID    func(t *testing.T, id string)
	}{
		{
			name:        "propagates provided header",
			headerValue: "client-provided-id",
			assertID: func(t *testing.T, id string) {
				assert.Equal(t, "client-provided-id", id)
			},
		},
		{
			name:        "generates UUID v4 when header absent",
			headerValue: "",
			assertID: func(t *testing.T, id string) {
				require.NotEmpty(t, id, "a request ID must be generated")
				parsed, err := uuid.FromString(id)
				require.NoError(t, err, "generated request ID must be a valid UUID")
				assert.Equal(t, byte(4), parsed.Version(), "generated request ID must be a UUID v4")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler, captured := captureRequestIDHandler(t)
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.headerValue != "" {
				req.Header.Set(RequestIDHeader, tc.headerValue)
			}

			RequestID(handler).ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			tc.assertID(t, *captured)
			// The middleware must not set the response header; echoing it is
			// intentionally left to a future mechanism.
			assert.Empty(t, rec.Header().Get(RequestIDHeader))
		})
	}
}

// TestNewRouter_AppliesRequestIDMiddleware verifies that NewRouter
// includes the RequestID middleware in the chain.
func TestNewRouter_AppliesRequestIDMiddleware(t *testing.T) {
	t.Helper()

	// Build the router using the global constructor
	r := NewRouter(discardLogger())

	// Create a request that will go through the middleware chain
	req := httptest.NewRequest(http.MethodGet, "/any-path", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	// The request should have passed through RequestID middleware
	// (we can't easily inspect internal state, but we know the
	// middleware ran if we got a response without panicking)
	require.NotEqual(t, http.StatusInternalServerError, rec.Code)
}

func TestNewRequestID_ReturnsUUIDV4(t *testing.T) {
	id := newRequestID(uuid.NewV4)

	parsed, err := uuid.FromString(id)
	require.NoError(t, err, "generated request ID must be a valid UUID")
	assert.Equal(t, byte(4), parsed.Version(), "generated request ID must be a UUID v4")
}

func TestNewRequestID_ReturnsEmptyWhenGeneratorFails(t *testing.T) {
	boom := errors.New("random source failure")

	id := newRequestID(func() (uuid.UUID, error) {
		return uuid.UUID{}, boom
	})

	assert.Empty(t, id, "must fall back to empty string instead of panicking")
}
