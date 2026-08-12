package metrics

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/muriiloandrade/finsplitter/pkg/logctx"
	"github.com/muriiloandrade/finsplitter/pkg/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/sdk/metric"
)

// NewMeterProvider creates and configures a new OpenTelemetry MeterProvider with OTLP HTTP exporter.
// Returns the meter provider, a shutdown function that should be called when the application exits,
// and an error if initialization fails.
//
// The meter provider is automatically set as the global provider and can be accessed via otel.Meter().
func NewMeterProvider(ctx context.Context, opts telemetry.Options) (telemetry.ShutdownFunc, error) {
	logger := logctx.FromCtx(ctx)

	// Create resource with service information
	res, err := telemetry.NewResource(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create OTLP HTTP exporter
	exporter, err := createExporter(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create meter provider with periodic reader
	mp := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exporter,
			metric.WithInterval(opts.ExportInterval()),
		)),
		metric.WithResource(res),
	)

	// Set global meter provider
	otel.SetMeterProvider(mp)

	logger.InfoContext(ctx, "meter provider initialized",
		slog.String("service", opts.ServiceName()),
		slog.String("version", opts.ServiceVersion()),
		slog.String("environment", opts.Environment()),
		slog.String("exporter", opts.ExporterURL()),
		slog.Bool("insecure", opts.Insecure()),
		slog.Duration("export_interval", opts.ExportInterval()),
		slog.Duration("export_timeout", opts.ExporterTimeout()),
	)

	// Return provider and shutdown function
	return mp.Shutdown, nil
}

// createExporter creates an OTLP HTTP exporter with the given options.
func createExporter(ctx context.Context, opts telemetry.Options) (metric.Exporter, error) {
	const firstRetryIn = 500 * time.Millisecond
	exporterOpts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(opts.ExporterURL()),
		otlpmetrichttp.WithTimeout(opts.ExporterTimeout()),
		otlpmetrichttp.WithCompression(otlpmetrichttp.GzipCompression),
		otlpmetrichttp.WithRetry(otlpmetrichttp.RetryConfig{
			InitialInterval: firstRetryIn,
		}),
	}

	if opts.Insecure() {
		exporterOpts = append(exporterOpts, otlpmetrichttp.WithInsecure())
	}

	return otlpmetrichttp.New(ctx, exporterOpts...)
}
