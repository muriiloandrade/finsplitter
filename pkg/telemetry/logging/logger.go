package logging

import (
	"context"
	"io"
	"log/slog"
	"os"

	"github.com/muriiloandrade/finsplitter/app/config"
	slogctx "github.com/veqryn/slog-context"
)

func NewContextWithLogger(ctx context.Context, cfg config.Config, w io.Writer) context.Context {
	defaultAttrs := []slog.Attr{
		slog.Group(
			"application",
			slog.String("service", cfg.App.Name),
			slog.Int("port", cfg.App.Port),
			slog.String("version", cfg.App.Version),
			slog.String("environment", cfg.Env.Name),
		),
		slog.Group(
			"build",
			slog.String("buildTime", cfg.App.BuildTime),
			slog.String("buildTag", cfg.App.BuildTag),
			slog.String("buildCommit", cfg.App.BuildCommit),
		),
	}

	var logHandler slog.Handler
	switch cfg.Env.LogFormat {
	case "text":
		logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelDebug,
		}).WithAttrs(defaultAttrs)
	case "json":
		logHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelDebug,
		}).WithAttrs(defaultAttrs)
	}

	customHandler := slogctx.NewHandler(logHandler, nil)

	logger := slog.New(customHandler)

	slog.SetDefault(logger)

	mainCtx := slogctx.NewCtx(ctx, slog.Default())

	return mainCtx
}
