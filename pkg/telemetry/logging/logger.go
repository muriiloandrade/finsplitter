package logging

import (
	"context"
	"log/slog"
	"os"

	"github.com/dusted-go/logging/prettylog"
	"github.com/muriiloandrade/finsplitter/internal/config"
	"github.com/muriiloandrade/finsplitter/pkg/logctx"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/sdk/log"
)

// NewContextWithLogger creates a new context with a configured logger.
// If otelLoggerProvider is provided, logs will be sent to both local output and OTel collector.
// If otelLoggerProvider is nil, only local logging is configured (useful for tests and local dev without OTel).
func NewContextWithLogger(
	ctx context.Context,
	cfg config.Config,
	otelLoggerProvider *log.LoggerProvider,
) context.Context {
	defaultAttrs := []slog.Attr{
		slog.Group(
			"application",
			slog.String("service", cfg.App.Name),
			slog.Int("port", cfg.App.Port),
			slog.String("version", cfg.App.Version),
			slog.String("environment", cfg.Env.Name),
		),
	}

	undefString := "undefined"
	if cfg.App.BuildCommit != undefString || cfg.App.BuildTag != undefString ||
		cfg.App.BuildTime != undefString {
		defaultAttrs = append(defaultAttrs, slog.Group(
			"build",
			slog.String("buildTime", cfg.App.BuildTime),
			slog.String("buildTag", cfg.App.BuildTag),
			slog.String("buildCommit", cfg.App.BuildCommit),
		))
	}

	var logHandler slog.Handler
	switch cfg.Env.LogFormat {
	case "text":
		logHandler = prettylog.NewHandler(&slog.HandlerOptions{
			Level:       slog.LevelDebug,
			AddSource:   false,
			ReplaceAttr: nil,
		}).WithAttrs(defaultAttrs)
	case "json":
	default:
		logHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelDebug,
		}).WithAttrs(defaultAttrs)
	}

	// Wrap with CustomHandler for context value extraction
	var finalHandler slog.Handler = &CustomHandler{Handler: logHandler}

	// If OTel logger provider is provided, create multi-handler to send logs to both local and OTel
	if otelLoggerProvider != nil {
		otelHandler := otelslog.NewHandler(
			cfg.App.Name,
			otelslog.WithLoggerProvider(otelLoggerProvider),
			otelslog.WithVersion(cfg.App.Version),
		)
		finalHandler = NewMultiHandler(finalHandler, otelHandler)
	}

	logger := slog.New(finalHandler)

	slog.SetDefault(logger)

	mainCtx := logctx.NewCtx(ctx, slog.Default())

	return mainCtx
}
