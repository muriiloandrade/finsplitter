package http

import (
	"net/http"
	"uuid"

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
			requestID = newRequestID(uuid.New)
		}

		ctx := logctx.WithRequestID(r.Context(), requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// newRequestID formats the UUID produced by newID as a string. The generator is
// injected so tests can pin the output; production passes uuid.New (UUID v4).
//
// The standard library uuid package cannot report a generation error the way
// gofrs/uuid did: uuid.New reads from crypto/rand, which crashes the program
// irrecoverably if the system source fails (it is not a recoverable panic).
// There is therefore no fallback ID to return — an entropy failure is fatal by
// upstream design, and the previous empty-string fallback is unreachable.
func newRequestID(newID func() uuid.UUID) string {
	return newID().String()
}
