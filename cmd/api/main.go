package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/jackc/pgx/v5/pgxpool"
	cbUCs "github.com/muriiloandrade/finsplitter/internal/app/usecases/card-brand"
	"github.com/muriiloandrade/finsplitter/internal/config"
	_http "github.com/muriiloandrade/finsplitter/internal/gateways/http"
	v1 "github.com/muriiloandrade/finsplitter/internal/gateways/http/v1"
	cbHandler "github.com/muriiloandrade/finsplitter/internal/gateways/http/v1/card-brand"
	"github.com/muriiloandrade/finsplitter/internal/gateways/postgres"
	"github.com/muriiloandrade/finsplitter/internal/gateways/postgres/migrations"
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
	cli := humacli.New(func(hooks humacli.Hooks, _ *CLIOptions) {
		cfg := config.LoadEnv(BuildTag, BuildCommit, BuildTime)
		if cfg == nil {
			panic("failed to load config")
		}

		ctx := logging.NewContextWithLogger(context.Background(), *cfg)
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
			panic(
				fmt.Errorf("failed to connect to database: %w", err),
			) // Exit if database connection fails
		}

		// Run database migrations
		err = migrations.RunMigrations(ctx, migrations.MigrationOptions{
			MigrationsPath: "./internal/gateways/postgres/migrations",
			DBInstance:     dbPool,
			DBCfg:          cfg.DB,
		})
		if err != nil {
			panic(
				fmt.Errorf("failed to run database migrations: %w", err),
			) // Exit if migrations fail
		}

		pgTxManager := &postgres.TxManager{
			ConnPool: dbPool,
		}

		// Initialize the repositories
		cardBrandRepo := postgres.NewCardBrandRepository(pgTxManager)

		// Initialize the use cases
		getCardBrandUC := cbUCs.NewGetCardBrandByIDUC(cardBrandRepo)
		listCardBrandUC := cbUCs.NewListCardBrandUC(cardBrandRepo)
		createCardBrandUC := cbUCs.NewCreateCardBrandUC(cardBrandRepo, pgTxManager)
		updateCardBrandUC := cbUCs.NewUpdateCardBrandUC(cardBrandRepo, pgTxManager)
		deleteCardBrandUC := cbUCs.NewDeleteCardBrandUC(cardBrandRepo, pgTxManager)

		cardBrandAPI := cbHandler.API{
			GetCardBrandHandler: cbHandler.NewGetCardBrandHandler(getCardBrandUC).GetCardBrand,
			ListCardBrandsHandler: cbHandler.NewListCardBrandsHandler(
				listCardBrandUC,
			).ListCardBrands,
			CreateCardBrandHandler: cbHandler.NewCreateCardBrandHandler(
				createCardBrandUC,
			).CreateCardBrand,
			UpdateCardBrandHandler: cbHandler.NewUpdateCardBrandHandler(
				updateCardBrandUC,
			).UpdateCardBrand,
			DeleteCardBrandHandler: cbHandler.NewDeleteCardBrandHandler(
				deleteCardBrandUC,
			).DeleteCardBrand,
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
			err = server.ListenAndServe()
			if err != nil {
				logger.Error("Failed to start server", slog.Any("error", err))
			}
		})

		// Tell the CLI how to stop your server.
		hooks.OnStop(func() {
			// Give the server 5 seconds to gracefully shut down, then give up.
			const timeout = 5 * time.Second
			timeoutCtx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			defer dbPool.Close()
			if err = server.Shutdown(timeoutCtx); err != nil {
				logger.Error("Failed to stop server", slog.Any("error", err))
			}
			logger.InfoContext(timeoutCtx, "Server stopped")
		})
	})

	// Run the CLI. When passed no commands, it starts the server.
	cli.Run()
}
