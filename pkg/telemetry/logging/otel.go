package logging

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/muriiloandrade/finsplitter/pkg/telemetry"
	slogctx "github.com/veqryn/slog-context"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/sdk/log"
)

// NewLoggerProvider creates and configures a new OpenTelemetry LoggerProvider with OTLP HTTP exporter.
// Returns the logger provider, a shutdown function that should be called when the application exits,
// and an error if initialization fails.
//
// The logger provider is automatically set as the global provider and can be used with otelslog bridge.
func NewLoggerProvider(
	ctx context.Context,
	opts telemetry.Options,
) (*log.LoggerProvider, telemetry.ShutdownFunc, error) {
	logger := slogctx.FromCtx(ctx)

	// Create resource with service information
	res, err := telemetry.NewResource(ctx, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create OTLP HTTP exporter
	exporter, err := createLogExporter(ctx, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create logger provider with batch processor
	lp := log.NewLoggerProvider(
		log.WithProcessor(log.NewBatchProcessor(exporter)),
		log.WithResource(res),
	)

	logger.InfoContext(ctx, "logger provider initialized",
		slog.String("service", opts.ServiceName()),
		slog.String("version", opts.ServiceVersion()),
		slog.String("environment", opts.Environment()),
		slog.String("exporter", opts.ExporterURL()),
		slog.Bool("insecure", opts.Insecure()),
	)

	// Return provider and shutdown function
	return lp, lp.Shutdown, nil
}

// createLogExporter creates an OTLP HTTP log exporter with the given options.
func createLogExporter(ctx context.Context, opts telemetry.Options) (log.Exporter, error) {
	const firstRetryIn = 500 * time.Millisecond
	exporterOpts := []otlploghttp.Option{
		otlploghttp.WithEndpoint(opts.ExporterURL()),
		otlploghttp.WithTimeout(opts.ExporterTimeout()),
		otlploghttp.WithCompression(otlploghttp.GzipCompression),
		otlploghttp.WithRetry(otlploghttp.RetryConfig{
			InitialInterval: firstRetryIn,
		}),
	}

	if opts.Insecure() {
		exporterOpts = append(exporterOpts, otlploghttp.WithInsecure())
	}

	return otlploghttp.New(ctx, exporterOpts...)
}
