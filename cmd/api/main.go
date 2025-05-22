package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/muriiloandrade/finsplitter/app/config"
	"github.com/muriiloandrade/finsplitter/app/domain/usecases"
	_http "github.com/muriiloandrade/finsplitter/app/gateways/http"
	v1 "github.com/muriiloandrade/finsplitter/app/gateways/http/v1"
	cardbrand "github.com/muriiloandrade/finsplitter/app/gateways/http/v1/card-brand"
	"github.com/muriiloandrade/finsplitter/app/gateways/postgres"
	"github.com/muriiloandrade/finsplitter/app/gateways/postgres/migrations"
	"github.com/muriiloandrade/finsplitter/pkg/telemetry/logging"
	slogctx "github.com/veqryn/slog-context"
)

// CLIOptions for the CLI.
type CLIOptions struct{}

var (
	BuildCommit = "undefined"
	BuildTag    = "undefined"
	BuildTime   = "undefined"
)

func main() {
	// Create a CLI app
	cli := humacli.New(func(hooks humacli.Hooks, options *CLIOptions) {
		cfg := config.LoadEnv(BuildTag, BuildCommit, BuildTime)
		if cfg == nil {
			panic("failed to load config")
		}

		ctx := logging.NewContextWithLogger(context.Background(), *cfg, os.Stdout)
		logger := slogctx.FromCtx(ctx)

		poolCfg, err := postgres.NewPoolConfig(cfg.DB)
		if err != nil {
			panic(fmt.Errorf("failed to create pool config: %w", err)) // Exit if pool config fails
		}

		dbPool, err := pgxpool.NewWithConfig(
			ctx,
			poolCfg,
		)
		if err != nil {
			panic(fmt.Errorf("failed to connect to database: %w", err)) // Exit if database connection fails
		}

		// Run database migrations
		err = migrations.RunMigrations(ctx, migrations.MigrationOptions{
			MigrationsPath: "./app/gateways/postgres/migrations",
			DbInstance:     dbPool,
			DbCfg:          cfg.DB,
		})
		if err != nil {
			panic(fmt.Errorf("failed to run database migrations: %w", err)) // Exit if migrations fail
		}

		pgTxManager := &postgres.TxManager{
			ConnPool: dbPool,
		}

		// Initialize the repositories
		cardBrandRepo := postgres.NewCardBrandRepository(pgTxManager)
		cardBrandUC := usecases.NewListCardBrandUC(cardBrandRepo)
		cardBrandAPI := cardbrand.API{
			ListCardBrandsHandler: cardbrand.NewListCardBrandsHandler(cardBrandUC).ListCardBrands,
		}

		router := _http.NewRouter(logger)

		apiV1 := v1.API{
			LivenessHandler:  v1.LivenessHandler(),
			ReadinessHandler: v1.ReadinessHandler(),
			CardBrandAPI:     cardBrandAPI,
			Logger:           logger,
		}

		apiV1.Routes(router)

		// Create the HTTP server.
		server := http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.App.Port),
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
			defer dbPool.Close()
			server.Shutdown(ctx)
		})
	})

	// Run the CLI. When passed no commands, it starts the server.
	cli.Run()
}
