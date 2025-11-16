package telemetry

import (
	"context"
	"errors"
	"log/slog"

	slogctx "github.com/veqryn/slog-context"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

// NewResource creates a new OpenTelemetry resource with service information.
// It automatically detects process, OS, container, and host information.
func NewResource(ctx context.Context, opts Options) (*resource.Resource, error) {
	logger := slogctx.FromCtx(ctx)

	// Create resource with automatic detection and service attributes
	res, err := resource.New(ctx,
		resource.WithProcess(),
		resource.WithOS(),
		resource.WithContainer(),
		resource.WithHost(),
		resource.WithAttributes(
			semconv.ServiceName(opts.ServiceName()),
			semconv.ServiceVersion(opts.ServiceVersion()),
			semconv.DeploymentEnvironmentName(opts.Environment()),
		),
	)

	// Handle partial resource errors (non-fatal)
	if errors.Is(err, resource.ErrPartialResource) || errors.Is(err, resource.ErrSchemaURLConflict) {
		logger.WarnContext(ctx, "non-fatal resource detection error",
			slog.Any("error", err),
			slog.String("service", opts.ServiceName()),
		)
	} else if err != nil {
		return nil, err
	}

	// Merge with default resource (adds telemetry SDK info)
	return resource.Merge(
		resource.Default(),
		res,
	)
}
