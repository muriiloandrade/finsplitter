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
func NewCtx(parent context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(parent, ctxKey{}, logger)
}

// FromCtx returns the logger stored in ctx, or slog.Default() if none is stored.
func FromCtx(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}
