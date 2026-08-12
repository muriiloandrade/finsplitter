package http

import (
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/muriiloandrade/finsplitter/pkg/logctx"
)

// RequestIDHeader is the name of the HTTP header that carries the request ID.
const RequestIDHeader = "X-Request-ID"

// RequestID is a middleware that reads the X-Request-ID header, generates a
// UUID v4 when the header is absent, and stores the ID in the request context
// under logctx.RequestIDKey so log lines for the request can include it.
//
// A custom middleware is used instead of chi's middleware.RequestID because
// the built-in generates hostname/counter IDs (not UUIDs) and exposes no hook
// to override the ID generation function — only RequestIDHeader is
// configurable. A UUID v4 in every log line is a requirement, so the
// generation logic is owned here.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			requestID = newRequestID()
		}

		ctx := logctx.WithRequestID(r.Context(), requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// newRequestID generates a UUID v4. On the astronomically unlikely failure of
// the random source it falls back to the empty string rather than panicking,
// so a request is never dropped because an ID could not be minted.
func newRequestID() string {
	id, err := uuid.NewV4()
	if err != nil {
		return ""
	}
	return id.String()
}
