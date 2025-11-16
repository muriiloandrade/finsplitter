package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// StartSpan starts a new span with the given name and options.
// It returns the new context containing the span and the span itself.
// The span should be ended by calling span.End() when done.
func StartSpan(
	ctx context.Context,
	tracerName, spanName string,
	opts ...trace.SpanStartOption,
) (context.Context, trace.Span) {
	tracer := otel.Tracer(tracerName)
	return tracer.Start(ctx, spanName, opts...) //nolint:spancheck // Span is ended by the caller
}

// SpanFromContext returns the current span from the context.
// If there is no span in the context, it returns a no-op span.
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}
