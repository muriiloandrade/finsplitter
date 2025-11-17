package logging

import (
	"context"
	"log/slog"
)

// MultiHandler forwards log records to multiple handlers.
// This allows logs to be sent to both local output (stdout/prettylog) and OTel collector.
type MultiHandler struct {
	handlers []slog.Handler
}

// NewMultiHandler creates a new MultiHandler that forwards records to all provided handlers.
func NewMultiHandler(handlers ...slog.Handler) *MultiHandler {
	return &MultiHandler{handlers: handlers}
}

// Enabled returns true if any of the handlers are enabled for the given level.
func (h *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

// Handle forwards the record to all handlers.
func (h *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, r.Level) {
			if err := handler.Handle(ctx, r.Clone()); err != nil {
				return err
			}
		}
	}
	return nil
}

// WithAttrs returns a new MultiHandler with the given attributes added to all handlers.
func (h *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithAttrs(attrs)
	}
	return &MultiHandler{handlers: newHandlers}
}

// WithGroup returns a new MultiHandler with the given group added to all handlers.
func (h *MultiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithGroup(name)
	}
	return &MultiHandler{handlers: newHandlers}
}
