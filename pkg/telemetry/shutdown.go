package telemetry

import "context"

// ShutdownFunc is a function that shuts down an OpenTelemetry provider.
// It should be called when the application exits to ensure all telemetry data is flushed.
type ShutdownFunc func(context.Context) error
