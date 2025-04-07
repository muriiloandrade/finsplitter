package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/muriiloandrade/finsplitter/app/config"
	_http "github.com/muriiloandrade/finsplitter/app/gateways/http"
	v1 "github.com/muriiloandrade/finsplitter/app/gateways/http/v1"
	"github.com/muriiloandrade/finsplitter/pkg/telemetry/logging"
	slogctx "github.com/veqryn/slog-context"
)

// CLIOptions for the CLI.
type CLIOptions struct {
	Port int `help:"Port to listen on" short:"p" default:"3033"`
}

func main() {
	// Create a CLI app
	cli := humacli.New(func(hooks humacli.Hooks, options *CLIOptions) {
		ctx := logging.NewContextWithLogger(context.Background(), config.Config{
			App: config.Application{
				Name:    "finsplitter",
				Version: "0.0.1",
			},
			Env: config.Environment{
				Name:      "local",
				LogFormat: "text",
			},
		}, os.Stdout)
		logger := slogctx.FromCtx(ctx)

		router := _http.NewRouter(logger)

		apiV1 := v1.API{
			LivenessHandler:  v1.LivenessHandler(),
			ReadinessHandler: v1.ReadinessHandler(),
			Logger:           logger,
		}

		apiV1.Routes(router)

		// Create the HTTP server.
		server := http.Server{
			Addr:    fmt.Sprintf(":%d", options.Port),
			Handler: router,
		}

		// Tell the CLI how to start your router.
		hooks.OnStart(func() {
			logger.Info("Starting server...")
			server.ListenAndServe()
		})

		// Tell the CLI how to stop your server.
		hooks.OnStop(func() {
			// Give the server 5 seconds to gracefully shut down, then give up.
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			server.Shutdown(ctx)
		})
	})

	// Run the CLI. When passed no commands, it starts the server.
	cli.Run()
}
