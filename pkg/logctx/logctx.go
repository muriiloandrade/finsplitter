// Package logctx provides context-based logger retrieval and storage.
//
// The standard library's log/slog package does not provide FromContext/NewContext
// helpers, so this small package fills that gap using a private context key.
// It mirrors the API previously provided by github.com/veqryn/slog-context.
package logctx

import (
	"context"
	"log/slog"
)

type ctxKey struct{}

// NewCtx returns a copy of parent with the given logger stored in it.
// A nil parent falls back to context.Background(), matching the semantics
// of the veqryn/slog-context API this package replaces.
func NewCtx(parent context.Context, logger *slog.Logger) context.Context {
	if parent == nil {
		parent = context.Background()
	}
	return context.WithValue(parent, ctxKey{}, logger)
}

// FromCtx returns the logger stored in ctx, or slog.Default() if none is stored.
// A nil ctx is treated as empty and returns slog.Default(), matching the
// semantics of the veqryn/slog-context API this package replaces.
func FromCtx(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return slog.Default()
	}
	if logger, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}
