package logctx

import "context"

// Key to use when setting the request ID.
type ctxKeyRequestID int

// RequestIDKey is the key that holds the request ID in a request context.
const RequestIDKey ctxKeyRequestID = 0

// WithRequestID returns a copy of parent with the request ID stored in it.
func WithRequestID(parent context.Context, requestID string) context.Context {
	return context.WithValue(parent, RequestIDKey, requestID)
}

// GetRequestID returns the request ID from the given context if one is present.
// Returns the empty string if a request ID cannot be found.
func GetRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}
