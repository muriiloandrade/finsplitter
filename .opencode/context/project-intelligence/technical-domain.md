<!-- Context: project-intelligence/technical | Priority: critical | Version: 1.2 | Updated: 2026-02-16 -->

# Technical Domain

**Purpose**: Tech stack, architecture, and development patterns for finsplitter.
**Last Updated**: 2026-02-16

## Quick Reference
**Update Triggers**: Tech stack changes | New patterns | Architecture decisions
**Audience**: Developers, AI agents

## Primary Stack
| Layer | Technology | Version | Rationale |
|-------|-----------|---------|-----------|
| Language | Go | 1.26 generics |
| Framework | Huma v2 | 2.35 | REST | Modern Go with API + OpenAPI auto-generation |
| Database | PostgreSQL | latest | Primary data store |
| DB Driver | pgx | 5.8 | High-performance PostgreSQL driver |
| Observability | OpenTelemetry | 1.40 | Distributed tracing, metrics, logs |
| Config | ardanlabs/conf | 3.10 | Environment-based configuration |
| SQL Gen | sqlc | 1.29 | Type-safe SQL queries |
| Mockery | v3.5.3 | Docker | Interface mock generation |

## Architecture
**Pattern**: Clean Architecture (Ports & Adapters)

```
internal/
├── config/        # Configuration loading
├── domain/        # Entities, errors, interfaces
│   └── errs/     # Domain errors (ErrCardBrandAlreadyExists)
├── app/
│   ├── ports/    # Repository interfaces
│   └── usecases/ # Business logic (card-brand/create_card_brand.go)
├── gateways/
│   ├── http/v1/  # HTTP handlers (Huma v2)
│   └── postgres/ # DB implementation (pgx, sqlc)
cmd/api/main.go   # DI wiring
```

## Make Commands
```bash
make start-dev        # Infra + app with hot reload
make start-debug      # Start with delve debugging
make start-infra      # Only database/infrastructure
make stop-dev         # Stop all containers
make generate         # Run all generators (sqlc + mocks)
make generate-sqlc   # Generate DB code from queries/*.sql
make generate-mocks  # Generate mocks for ports/ and domain/
make new-migration name=create_table_foo  # Timestamp migration files
make migrate-up      # Apply pending migrations
make test            # Run go test
make code-check      # Format + lint (golangci-lint)
```

## Code Patterns

### API Handler (Huma v2)
```go
func (h CreateCardBrandHandler) CreateCardBrand(
    ctx context.Context, input *CreateCardBrandRequest,
) (*CreateCardBrandResponse, error) {
    logger := slogctx.FromCtx(ctx)
    brand, err := h.UseCase.CreateCardBrand(ctx, input.Body.Name)
    if errors.Is(err, errs.ErrCardBrandAlreadyExists) {
        return nil, huma.Error409Conflict(err.Error())
    }
    // ... handle errors → huma.Error4xxYYY()
}
```

### Use Case with Transaction
```go
func (uc *CreateCardBrandUC) CreateCardBrand(ctx context.Context, name string) (*entity.CardBrand, error) {
    var result *entity.CardBrand
    err := uc.tx.WithTx(ctx, func(ctx context.Context) error {
        r, err := uc.repo.CreateCardBrand(ctx, name)
        result = r
        return err
    })
    return result, err
}
```

### SQLC Query Pattern
```sql
-- name: CreateCardBrand :one
INSERT INTO card_brand (name) VALUES ($1) RETURNING *;
```

### OTel Functional Options
```go
opts, _ := telemetry.NewOptions(
    telemetry.WithServiceName("finsplitter"),
    telemetry.WithExporterURL("otel-collector:4318"),
    telemetry.WithInsecure(true),
)
tracer, _, _ := tracing.NewTracerProvider(ctx, opts, 1.0)
```

## Naming Conventions
| Type | Convention | Example |
|------|-----------|---------|
| Files | snake_case | `create_card_brand.go` |
| Packages | lowercase | `cardbrand`, `usecases` |
| Exported Functions | PascalCase | `GetCardBrandHandler` |
| Interfaces | PascalCase + suffix | `Repository`, `UseCase` |
| Database Tables | snake_case | `card_brands` |

## Code Standards
- Use `errors.Is()` for error checking (not equality)
- Domain errors in `internal/domain/errs/errs.go`
- Map DB errors via `pgerrcode.IsIntegrityConstraintViolation()`
- Use `slogctx.FromCtx(ctx)` for structured logging
- Use sqlc for parameterized queries (safe from SQL injection)
- Configuration via `ardanlabs/conf` with env vars
- Test files `*_test.go` in same package with testify/mock
- UUIDs for entity IDs, auto `last_modified_date` via trigger

## Security Requirements
- Validate all input via Huma schema tags
- Parameterized queries (sqlc generates)
- Environment-based secrets via `conf`
- SSL/TLS for DB (`PG_SSL_MODE=require`)
- No sensitive data in logs
- Proper error messages (don't leak internals)

## Adding New Feature (Step-by-Step)
1. `make new-migration name=create_table_foo`
2. Write SQL in `sqlc/queries/foo.sql`
3. `make generate-sqlc` → generates Go code
4. Define entity in `domain/entity/foo.go`
5. Define ports in `app/ports/foo_repo.go`
6. Implement in `gateways/postgres/foo.go`
7. Create use cases in `app/usecases/foo/`
8. Create handlers in `gateways/http/v1/foo/`
9. Wire in `cmd/api/main.go`
10. `make generate-mocks` + write tests

## 📂 Codebase References
- **Config**: `internal/config/config.go`
- **Domain Errors**: `internal/domain/errs/errs.go`
- **Telemetry**: `pkg/telemetry/` (options, tracing, metrics, logging)
- **Migration**: `internal/gateways/postgres/migrations/`
- **SQL Queries**: `internal/gateways/postgres/sqlc/queries/`

## Related Context Files
- [golang-patterns.md](./golang-patterns.md) - Go patterns, testing, security
- [database-patterns.md](./database-patterns.md) - PostgreSQL, sqlc, migrations
- [docker-patterns.md](./docker-patterns.md) - Containerization patterns
- [workflow.md](./workflow.md) - Development workflow & quality
