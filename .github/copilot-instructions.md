# Finsplitter - AI Coding Agent Instructions

## Architecture Overview

**Clean Architecture Pattern**: The codebase follows hexagonal/ports-and-adapters architecture with strict layer separation:
- `internal/domain/entity`: Core business entities (e.g., `CardBrand`, `Person`, `Card`)
- `internal/app/ports`: Repository interfaces defining data contracts
- `internal/app/usecases`: Business logic orchestration (e.g., `card-brand/create_card_brand.go`)
- `internal/gateways/postgres`: Database implementation using pgx/v5 and SQLC
- `internal/gateways/http/v1`: HTTP handlers using Huma v2 framework

**Key Architectural Decisions**:
- All use cases requiring writes accept `domain.Transactioner` for transaction management
- Context-aware database operations via `TxManager.WithTx()` - checks `domain.HasTX(ctx)` to prevent nested transactions
- Repository methods accept context and extract transactions via type assertion from context values
- SQLC generates type-safe database access code from SQL queries

## Critical Workflows

### Development Workflow
```bash
make start-dev        # Start infra + app with hot reload
make start-debug      # Start with debugging enabled (delve)
make stop-dev         # Stop all containers
make start-infra      # Only start database/infrastructure
```

**Docker profiles**: `infra`, `backend`, `debug` control service composition via docker compose

### Code Generation
```bash
make generate         # Run all generators (SQLC + mocks)
make generate-sqlc    # Generate postgres code from internal/gateways/postgres/sqlc/queries/*.sql
make generate-mocks   # Generate mocks via mockery for interfaces in ports/ and domain/
```

**SQLC Pattern**: Write SQL in `internal/gateways/postgres/sqlc/queries/*.sql` with annotations:
```sql
-- name: CreateCardBrand :one
INSERT INTO card_brand (name) VALUES ($1) RETURNING *;
```

### Database Migrations
```bash
make new-migration name=create_table_foo  # Creates timestamped up/down SQL files
make migrate-up                           # Apply all pending migrations
make migrate-down-1                       # Rollback last migration
```

Migrations live in `internal/gateways/postgres/migrations/` and run automatically on app startup via `migrations.RunMigrations()` in `cmd/api/main.go`.

### Testing & Quality
```bash
make test             # Run all tests with go test
make code-check       # Format + lint (golangci-lint via Docker)
make format           # Format code with golangci-lint fmt
```

**Test Pattern** (see `internal/app/usecases/card-brand/*_test.go`):
- Use testify/mock for repository and transaction mocks
- Table-driven tests with `repoSetup` and `txSetup` functions
- Mock transaction execution with `tx.EXPECT().WithTx()` that calls the callback

## Project-Specific Conventions

### Handler Pattern
HTTP handlers follow a 3-layer structure:
1. **Handler struct** (e.g., `CreateCardBrandHandler`) - wraps use case
2. **Request/Response structs** - define Huma contracts with validation tags
3. **Handler method** - extracts logger from context, calls use case, maps errors to HTTP codes

Example from `internal/gateways/http/v1/card-brand/create_card_brand.go`:
```go
func (h CreateCardBrandHandler) CreateCardBrand(
    ctx context.Context,
    input *CreateCardBrandRequest,
) (*CreateCardBrandResponse, error) {
    logger := slogctx.FromCtx(ctx)
    brand, err := h.UseCase.CreateCardBrand(ctx, input.Body.Name)
    if errors.Is(err, errs.ErrCardBrandAlreadyExists) {
        return nil, huma.Error409Conflict(err.Error())
    }
    // ... handle other errors
}
```

### Error Handling
- Domain errors defined in `internal/domain/errs/errs.go` (e.g., `ErrCardBrandAlreadyExists`)
- Repositories detect postgres errors and map to domain errors using `pgerrcode.IsIntegrityConstraintViolation()`
- Handlers convert domain errors to HTTP status codes via `huma.Error4xxYYY()` helpers

### Transaction Management
Use cases requiring atomicity wrap operations in `tx.WithTx()`:
```go
err := uc.tx.WithTx(ctx, func(ctx context.Context) error {
    result, err := uc.repo.CreateCardBrand(ctx, name)
    // ... additional operations in same transaction
    return err
})
```

Repositories extract connection from context in `querier` interface methods.

### Logging
- Structured logging via `log/slog` with context propagation (`slogctx.FromCtx(ctx)`)
- Initialize logger with app metadata in `cmd/api/main.go` via `logging.NewContextWithLogger()`
- Support "text" (prettylog) and "json" formats controlled by `LOG_FORMAT` env var
- Optional OpenTelemetry log export via `MultiHandler` pattern (dual local + OTel output)

### OpenTelemetry Observability
**Package Structure** (`pkg/telemetry`):
- `options.go`: Shared `Options` struct with **Functional Options Pattern** for all providers
  - `WithServiceName(name)` - Required, validates non-empty
  - `WithServiceVersion(version)` - Defaults to "local"
  - `WithEnvironment(env)` - Defaults to "development"
  - `WithExporterURL(url)` - Required, validates non-empty
  - `WithInsecure(bool)` - Controls TLS (true for dev, false for prod)
  - `WithExporterTimeout(duration)` - Defaults to 10 seconds
  - `WithExportInterval(duration)` - Defaults to 60 seconds (for metrics)
  - `NewOptions(opts...)` - Creates Options with validation and defaults applied
- `resource.go`: Shared `NewResource()` function for resource detection (process, OS, container, host)
- `shutdown.go`: Shared `ShutdownFunc` type alias
- `tracing/`: `NewTracerProvider(ctx, opts, samplerRatio)` - traces with OTLP HTTP exporter
- `metrics/`: `NewMeterProvider(ctx, opts)` - metrics with periodic reader (uses opts.ExportInterval())
- `logging/`: `NewLoggerProvider(ctx, opts)` - logs with batch processor

**Usage Example**:
```go
opts, err := telemetry.NewOptions(
    telemetry.WithServiceName("finsplitter"),
    telemetry.WithServiceVersion("v1.0.0"),
    telemetry.WithEnvironment("production"),
    telemetry.WithExporterURL("otel-collector:4318"),
    telemetry.WithInsecure(false),
    telemetry.WithExporterTimeout(10 * time.Second),
    telemetry.WithExportInterval(60 * time.Second),
)
if err != nil {
    // Handle validation error
}
tracerProvider, shutdown, err := tracing.NewTracerProvider(ctx, opts, 1.0)
meterProvider, shutdown, err := metrics.NewMeterProvider(ctx, opts)
```

**Initialization** (`cmd/api/main.go`):
- Private `initializeOpenTelemetry(ctx, cfg)` function creates all providers based on config flags
- Providers initialized early in main(), before logger context creation
- Shutdown functions called via defer to flush data on exit
- Global provider registration via `otel.SetTracerProvider()`, `otel.SetMeterProvider()`

**Instrumentation**:
- HTTP: `otelchi.Middleware()` in `internal/gateways/http/global_router.go`
- PostgreSQL: `otelpgx.NewTracer()` in `internal/gateways/postgres/pool.go`
- Logs: `otelslog` bridge via `MultiHandler` in `pkg/telemetry/logging/logger.go`

**Configuration** (`internal/config/config.go`):
- `OTEL_ENABLED`: Master switch for all telemetry
- `OTEL_ENABLE_{TRACES|METRICS|LOGS}`: Individual signal controls
- `OTEL_SAMPLER_RATIO`: Trace sampling ratio (0.0-1.0)
- `OTEL_EXPORTER_OTLP_ENDPOINT`: OTLP HTTP endpoint (default: localhost:4318)
- `OTEL_EXPORTER_TIMEOUT`: OTLP exporter timeout (default: 5s)
- `OTEL_EXPORT_INTERVAL`: Metrics export interval (default: 30s)

### Dependency Injection
Manual DI in `cmd/api/main.go`:
1. Initialize `pgxpool.Pool` and `TxManager`
2. Create repositories (e.g., `postgres.NewCardBrandRepository(pgTxManager)`)
3. Create use cases with repository + transaction dependencies
4. Create handlers with use case dependencies
5. Assemble into `v1.API` struct and register routes

### Configuration
All config via environment variables loaded in `internal/config/config.go`:
- `APP_*`: Application settings (port, name, version)
- `PG_*`: Database connection and pool settings
- `ENV_NAME`, `LOG_FORMAT`: Environment configuration

Build info injected via ldflags: `BuildCommit`, `BuildTag`, `BuildTime`

## Integration Points

### API Framework (Huma v2)
- OpenAPI 3.1 auto-generated at `/openapi` and docs at `/docs`
- Routes registered via `huma.Register(api, huma.Operation{...}, handlerFunc)`
- Chi router underneath (`humachi.New()`) with health checks at `/health/liveness` and `/health/readiness`

### Database (PostgreSQL)
- Connection pooling via pgx/v5 with configurable pool settings (`PG_MAX_CONNS`, etc.)
- UUID primary keys via `uuid-ossp` extension
- Automatic `last_modified_date` updates via trigger function `update_last_modified()`

### Tooling
- **Lefthook**: Pre-commit hooks run `make code-check` (see `lefthook.yml`)
- **Mockery v3**: Generates mocks via Docker (`vektra/mockery:v3.5.3`)
- **SQLC v1.29**: Type-safe SQL code generation via Docker (`sqlc/sqlc:1.29.0`)
- **golangci-lint v2.4**: Linting via Docker with cached build artifacts

## Adding New Features

To add a new entity (e.g., "Transaction"):
1. Create migration: `make new-migration name=create_table_transaction`
2. Write SQL queries in `internal/gateways/postgres/sqlc/queries/transaction.sql`
3. Run `make generate-sqlc` to generate Go code
4. Define entity in `internal/domain/entity/transaction.go`
5. Define ports in `internal/app/ports/transaction_repo.go`
6. Implement repository in `internal/gateways/postgres/transaction.go`
7. Create use cases in `internal/app/usecases/transaction/`
8. Create HTTP handlers in `internal/gateways/http/v1/transaction/`
9. Wire dependencies in `cmd/api/main.go`
10. Run `make generate-mocks` and write tests

## Build & Deployment

Multi-stage Dockerfile:
- `setup`: Download dependencies
- `builder`: Compile binary with ldflags
- `production`: Distroless image with only binary

Build variables set via `--build-arg`: `GIT_COMMIT`, `GIT_BUILD_TAG`, `BUILD_TIME`
