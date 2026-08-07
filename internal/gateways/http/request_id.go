package http

import (
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/muriiloandrade/finsplitter/pkg/logctx"
)

// RequestIDHeader is the name of the HTTP header that carries the request ID.
const RequestIDHeader = "X-Request-ID"

// RequestID is a middleware that reads the X-Request-ID header, generates a
// UUID v4 when the header is absent, stores the ID in the request context
// under logctx.RequestIDKey, and echoes it back in the response header.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.Must(uuid.NewV4()).String()
		}

		ctx := logctx.WithRequestID(r.Context(), requestID)
		w.Header().Set(RequestIDHeader, requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
