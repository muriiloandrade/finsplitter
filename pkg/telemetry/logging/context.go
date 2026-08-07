package logging

import (
	"context"
	"log/slog"

	"github.com/muriiloandrade/finsplitter/pkg/logctx"
)

type CustomHandler struct {
	slog.Handler
}

func (h *CustomHandler) Handle(ctx context.Context, r slog.Record) error {
	if requestID := logctx.GetRequestID(ctx); requestID != "" {
		r.AddAttrs(slog.String("request_id", requestID))
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
