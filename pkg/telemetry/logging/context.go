package logging

import (
	"context"
	"log/slog"
)

type CustomHandler struct {
	slog.Handler
}

func (h *CustomHandler) Handle(ctx context.Context, r slog.Record) error {
	keys := []string{"request_id"}

	for _, key := range keys {
		if value := ctx.Value(key); value != nil {
			r.AddAttrs(slog.Any(key, value))
		}
	}

	return h.Handler.Handle(ctx, r)
}

func (h *CustomHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.Handler.Enabled(ctx, level)
}

func (h *CustomHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &CustomHandler{h.Handler.WithAttrs(attrs)}
}

func (h *CustomHandler) WithGroup(name string) slog.Handler {
	return &CustomHandler{h.Handler.WithGroup(name)}
}
