package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	cbUCs "github.com/muriiloandrade/finsplitter/internal/app/usecases/card-brand"
	"github.com/muriiloandrade/finsplitter/internal/config"
	_http "github.com/muriiloandrade/finsplitter/internal/gateways/http"
	v1 "github.com/muriiloandrade/finsplitter/internal/gateways/http/v1"
	authHandler "github.com/muriiloandrade/finsplitter/internal/gateways/http/v1/auth"
	cbHandler "github.com/muriiloandrade/finsplitter/internal/gateways/http/v1/card-brand"
	profileHandler "github.com/muriiloandrade/finsplitter/internal/gateways/http/v1/profile"
	"github.com/muriiloandrade/finsplitter/internal/gateways/logto"
	"github.com/muriiloandrade/finsplitter/internal/gateways/postgres"
	"github.com/muriiloandrade/finsplitter/internal/gateways/postgres/migrations"
	"github.com/muriiloandrade/finsplitter/pkg/cache"
	"github.com/muriiloandrade/finsplitter/pkg/logctx"
	"github.com/muriiloandrade/finsplitter/pkg/telemetry"
	"github.com/muriiloandrade/finsplitter/pkg/telemetry/logging"
	"github.com/muriiloandrade/finsplitter/pkg/telemetry/metrics"
	"github.com/muriiloandrade/finsplitter/pkg/telemetry/tracing"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

// CLIOptions for the CLI.
type CLIOptions struct{}

//nolint:gochecknoglobals // These are set at build time using ldflags.
var (
	BuildCommit = "undefined"
	BuildTag    = "undefined"
	BuildTime   = "undefined"
)

const (
	gracefulShutdownTimeout = 5 * time.Second
	readHeaderTimeout       = 30 * time.Second
)

//nolint:gocognit // Main function is inherently complex due to DI wiring.
func main() {
	// Create a CLI app
	cli := humacli.New(func(hooks humacli.Hooks, _ *CLIOptions) {
		cfg := config.LoadEnv(BuildTag, BuildCommit, BuildTime)
		if cfg == nil {
			panic("failed to load config")
		}

		ctx := context.Background()

		// Initialize OpenTelemetry providers if enabled
		otelLoggerProvider, otelShutdown, otelErr := initializeOpenTelemetry(ctx, cfg)
		if otelErr != nil {
			panic(fmt.Errorf("failed to initialize OpenTelemetry: %w", otelErr))
		}

		ctx = logging.NewContextWithLogger(ctx, *cfg, otelLoggerProvider)
		logger := logctx.FromCtx(ctx)

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

		userRepo := postgres.NewUserRepository(pgTxManager)
		redisClient := newRedisClient(ctx, logger, cfg)
		logtoM2M := newLogtoM2MClient(logger, cfg)
		router := _http.NewRouter(logger)

		authMw := newAuthMiddleware(logger, cfg, userRepo, redisClient)

		apiV1 := v1.API{
			LivenessHandler:  v1.LivenessHandler(),
			ReadinessHandler: v1.ReadinessHandler(),
			CardBrandAPI:     newCardBrandAPI(pgTxManager),
			AuthAPI:          newAuthAPI(userRepo, logtoM2M),
			ProfileAPI:       newProfileAPI(userRepo, logtoM2M, authMw),
			AuthMiddleware:   authMw,
			Logger:           logger,
		}

		apiV1.Routes(router)

		// Create the HTTP server.
		server := http.Server{
			Addr:              fmt.Sprintf(":%d", cfg.App.Port),
			Handler:           router,
			ReadHeaderTimeout: readHeaderTimeout,
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
			timeoutCtx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
			defer cancel()
			defer dbPool.Close()
			if redisClient != nil {
				defer redisClient.Close()
			}
			if err = server.Shutdown(timeoutCtx); err != nil {
				logger.Error("Failed to stop server", slog.Any("error", err))
			}
			if otelShutdownErr := otelShutdown(ctx); otelShutdownErr != nil {
				logger.ErrorContext(ctx, "Failed to shutdown OpenTelemetry", slog.Any("error", otelShutdownErr))
			}
			logger.InfoContext(timeoutCtx, "Server stopped")
		})
	})

	// Run the CLI. When passed no commands, it starts the server.
	cli.Run()
}

// initializeOpenTelemetry initializes OpenTelemetry providers based on configuration.
// Returns the logger provider if logging is enabled, nil otherwise.
func initializeOpenTelemetry(
	ctx context.Context,
	cfg *config.Config,
) (*sdklog.LoggerProvider, func(context.Context) error, error) {
	if !cfg.OTel.Enabled {
		return nil, nil, nil
	}

	var shutdownFuncs []func(context.Context) error
	var err error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown := func(ctx context.Context) error {
		var shutdownErr error
		for _, fn := range shutdownFuncs {
			shutdownErr = errors.Join(shutdownErr, fn(ctx))
		}
		shutdownFuncs = nil
		return shutdownErr
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	opts, err := telemetry.NewOptions(
		telemetry.WithServiceName(cfg.OTel.ServiceName),
		telemetry.WithServiceVersion(cfg.App.Version),
		telemetry.WithEnvironment(cfg.Env.Name),
		telemetry.WithExporterURL(cfg.OTel.ExporterURL),
		telemetry.WithInsecure(cfg.OTel.Insecure),
		telemetry.WithExporterTimeout(cfg.OTel.ExporterTimeout),
		telemetry.WithExportInterval(cfg.OTel.ExportInterval),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create telemetry options: %w", err)
	}

	// Initialize tracer provider
	if cfg.OTel.EnableTraces {
		shutdownTracer, tpErr := tracing.NewTracerProvider(ctx, opts, cfg.OTel.SamplerRatio)
		if tpErr != nil {
			handleErr(tpErr)
			return nil, nil, fmt.Errorf("failed to initialize tracer provider: %w", tpErr)
		}
		shutdownFuncs = append(shutdownFuncs, shutdownTracer)
	}

	// Initialize meter provider
	if cfg.OTel.EnableMetrics {
		shutdownMeter, mpErr := metrics.NewMeterProvider(ctx, opts)
		if mpErr != nil {
			handleErr(mpErr)
			return nil, nil, fmt.Errorf("failed to initialize meter provider: %w", mpErr)
		}
		shutdownFuncs = append(shutdownFuncs, shutdownMeter)
	}

	// Initialize logger provider
	if cfg.OTel.EnableLogs {
		logProvider, shutdownLogger, lpErr := logging.NewLoggerProvider(ctx, opts)
		if lpErr != nil {
			handleErr(lpErr)
			return nil, nil, fmt.Errorf("failed to initialize logger provider: %w", lpErr)
		}
		shutdownFuncs = append(shutdownFuncs, shutdownLogger)
		return logProvider, shutdown, nil
	}

	return nil, nil, nil
}

// newRedisClient creates a cache client connected to Valkey/Redis.
// If the connection fails it logs a warning and returns nil, allowing the
// application to degrade gracefully (JWKS fetched on every request).
func newRedisClient(ctx context.Context, logger *slog.Logger, cfg *config.Config) *cache.Client {
	client, err := cache.New(ctx, cfg.Redis.URL)
	if err != nil {
		logger.WarnContext(ctx,
			"Redis unavailable, auth middleware will fetch JWKS on every request",
			slog.Any("error", err),
		)
		return nil
	}
	return client
}

func newLogtoM2MClient(logger *slog.Logger, cfg *config.Config) *logto.Client {
	return logto.NewClient(logto.Config{
		OIDCEndpoint:          cfg.Logto.OIDCEndpoint,
		ManagementBaseURL:     cfg.Logto.ManagementBaseURL,
		ManagementAPIResource: cfg.Logto.MgmtAPIResource,
		ClientID:              cfg.Logto.MgmtClientID,
		ClientSecret:          cfg.Logto.MgmtClientSecret,
		AppClientID:           cfg.Logto.AppClientID,
		AppClientSecret:       cfg.Logto.AppClientSecret,
		DeviceAppClientID:     cfg.Logto.DeviceAppClientID,
	}, logto.WithLogger(logger))
}

// newAuthMiddleware builds the JWT validation middleware.
func newAuthMiddleware(
	logger *slog.Logger,
	cfg *config.Config,
	userRepo ports.UserRepository,
	cacheClient *cache.Client,
) *authHandler.Middleware {
	return authHandler.NewMiddleware(
		cfg.Logto.OIDCEndpoint,
		cfg.Logto.OIDCIssuer,
		cfg.Logto.AppClientID,
		userRepo,
		logger,
		cacheClient,
	)
}

// newAuthAPI creates the auth handler API (register, me, device auth).
func newAuthAPI(userRepo ports.UserRepository, logtoClient *logto.Client) authHandler.API {
	return authHandler.NewAPI(userRepo, logtoClient)
}

// newProfileAPI creates the profile handler API (setup).
func newProfileAPI(
	userRepo ports.UserRepository,
	logtoM2M *logto.Client,
	claimsPr ports.ClaimsProvider,
) profileHandler.API {
	return profileHandler.NewAPI(userRepo, logtoM2M, claimsPr)
}

func newCardBrandAPI(pgTxManager *postgres.TxManager) cbHandler.API {
	cardBrandRepo := postgres.NewCardBrandRepository(pgTxManager)

	// Initialize the use cases
	getCardBrandUC := cbUCs.NewGetCardBrandByIDUC(cardBrandRepo)
	listCardBrandUC := cbUCs.NewListCardBrandUC(cardBrandRepo)
	createCardBrandUC := cbUCs.NewCreateCardBrandUC(cardBrandRepo, pgTxManager)
	updateCardBrandUC := cbUCs.NewUpdateCardBrandUC(cardBrandRepo, pgTxManager)
	deleteCardBrandUC := cbUCs.NewDeleteCardBrandUC(cardBrandRepo, pgTxManager)

	cardBrandAPI := cbHandler.API{
		GetCardBrandHandler: cbHandler.NewGetCardBrandHandler(&getCardBrandUC).GetCardBrand,
		ListCardBrandsHandler: cbHandler.NewListCardBrandsHandler(
			&listCardBrandUC,
		).ListCardBrands,
		CreateCardBrandHandler: cbHandler.NewCreateCardBrandHandler(
			&createCardBrandUC,
		).CreateCardBrand,
		UpdateCardBrandHandler: cbHandler.NewUpdateCardBrandHandler(
			&updateCardBrandUC,
		).UpdateCardBrand,
		DeleteCardBrandHandler: cbHandler.NewDeleteCardBrandHandler(
			&deleteCardBrandUC,
		).DeleteCardBrand,
	}
	return cardBrandAPI
}
