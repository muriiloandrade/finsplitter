package tracing

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/muriiloandrade/finsplitter/pkg/telemetry"
	slogctx "github.com/veqryn/slog-context"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const (
	// Default batch processor configuration values.
	defaultMaxQueueSize       = 2048
	defaultBatchTimeout       = 5 * time.Second
	defaultMaxExportBatchSize = 512
	defaultSamplerRatio       = 1.0
)

// NewTracerProvider creates and configures a new OpenTelemetry TracerProvider with OTLP HTTP exporter.
// Returns the tracer provider, a shutdown function that should be called when the application exits,
// and an error if initialization fails.
//
// The tracer provider is automatically set as the global provider and can be accessed via otel.Tracer().
func NewTracerProvider(
	ctx context.Context,
	opts telemetry.Options,
	samplerRatio float64,
) (telemetry.ShutdownFunc, error) {
	logger := slogctx.FromCtx(ctx)

	// Validate sampler ratio
	if err := validateSamplerRatio(samplerRatio); err != nil {
		return nil, fmt.Errorf("invalid sampler ratio: %w", err)
	}

	// Apply default sampler ratio
	if samplerRatio == 0.0 {
		samplerRatio = defaultSamplerRatio
	}

	// Create resource with service information
	res, err := telemetry.NewResource(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create OTLP HTTP exporter with timeout
	exporter, err := createExporter(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create tracer provider with batch span processor and optimized settings
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			// Batch span processor options for better performance
			sdktrace.WithMaxQueueSize(defaultMaxQueueSize),
			sdktrace.WithBatchTimeout(defaultBatchTimeout),
			sdktrace.WithMaxExportBatchSize(defaultMaxExportBatchSize),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(createSampler(samplerRatio)),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Set global propagator to W3C Trace Context and Baggage
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
		b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader|b3.B3SingleHeader)),
	))

	logger.InfoContext(ctx, "tracer provider initialized",
		slog.String("service", opts.ServiceName()),
		slog.String("version", opts.ServiceVersion()),
		slog.String("environment", opts.Environment()),
		slog.String("exporter", opts.ExporterURL()),
		slog.Bool("insecure", opts.Insecure()),
		slog.Float64("sampler_ratio", samplerRatio),
	)

	// Return provider and shutdown function
	return tp.Shutdown, nil
}

// createSampler creates a sampler based on the sampler ratio.
// Uses ParentBased sampler to respect parent sampling decisions.
func createSampler(samplerRatio float64) sdktrace.Sampler {
	// Always use ParentBased sampler to respect upstream sampling decisions
	if samplerRatio >= 1.0 {
		return sdktrace.ParentBased(sdktrace.AlwaysSample())
	}

	if samplerRatio <= 0.0 {
		return sdktrace.ParentBased(sdktrace.NeverSample())
	}

	// Use ParentBased TraceIDRatio sampler for probabilistic sampling
	return sdktrace.ParentBased(
		sdktrace.TraceIDRatioBased(samplerRatio),
	)
}

// validateSamplerRatio validates the sampler ratio parameter.
func validateSamplerRatio(samplerRatio float64) error {
	if samplerRatio < 0.0 || samplerRatio > 1.0 {
		return errors.New("samplerRatio must be between 0.0 and 1.0")
	}
	return nil
}

// createExporter creates an OTLP HTTP exporter with the given options.
func createExporter(ctx context.Context, opts telemetry.Options) (sdktrace.SpanExporter, error) {
	const firstRetryIn = 500 * time.Millisecond
	exporterOpts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(opts.ExporterURL()),
		otlptracehttp.WithTimeout(opts.ExporterTimeout()),
		otlptracehttp.WithCompression(otlptracehttp.GzipCompression),
		otlptracehttp.WithRetry(otlptracehttp.RetryConfig{
			InitialInterval: firstRetryIn,
		}),
	}

	if opts.Insecure() {
		exporterOpts = append(exporterOpts, otlptracehttp.WithInsecure())
	}

	return otlptracehttp.New(ctx, exporterOpts...)
}
