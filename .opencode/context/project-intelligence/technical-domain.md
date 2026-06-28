<!-- Context: project-intelligence/technical | Priority: critical | Version: 1.3 | Updated: 2026-06-27 -->

# Technical Domain

**Purpose**: Tech stack, architecture, and development patterns for finsplitter.
**Last Updated**: 2026-06-27

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
| JWT | lestrrat-go/jwx v4 + jwkfetch v4 | - | Logto-compatible JWT validation, JWKS fetching |
| Auth | Logto | latest | Identity provider & Management API |
| Cache | Valkey 9.1 + go-redis/v9 | 9.7 | Redis-compatible JWKS cache, rate limiting |
| Logging | veqryn/slog-context | 0.7 | Context-aware structured logging |
| Integration Testing | testcontainers-go | 0.43 | Real Valkey/Redis containers per test |
| Build | GOEXPERIMENT=jsonv2 | - | Required (jwx v4 depends on encoding/json/v2) |

## Architecture
**Pattern**: Clean Architecture (Ports & Adapters)

```
internal/
├── config/            # Env config (ardanlabs/conf)
├── domain/{entity,errs}  # Models + sentinel errors
├── app/
│   ├── ports/         # Interfaces: UserRepository, etc.
│   └── usecases/{auth,card-brand,profile}  # Business logic
├── gateways/
│   ├── http/v1/{auth,card-brand,profile}  # Handlers + routes
│   ├── logto/         # Logto M2M Management API client
│   └── postgres/      # pgx, sqlc, migrations
├── pkg/
│   ├── cache/         # Valkey/Redis + OTel instrumentation
│   ├── httpclient/    # Resty v3 wrapper
│   └── telemetry/     # OTel tracing, metrics, logs
cmd/api/main.go        # DI wiring
```

## Make Commands
- `make start-dev` — infra + hot reload
- `make new-migration name=create_table_foo` — timestamped migration
- `make generate` / `generate-sqlc` / `generate-mocks` — codegen
- `make test` / `make code-check` — test + format+lint

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

### Auth Middleware (jwx + JWKS Cache)
```go
// Middleware validates JWTs via Logto JWKS. Cached in Valkey with 15min TTL.
// Graceful degradation: cache failure → direct Logto fetch.
// Interface extracted for testability (jwkFetcher → mockjwkFetcher).
type Middleware struct {
    jwkClient jwkFetcher       // *jwkfetch.Client in production
    cache     *cache.Client    // Valkey-backed; nil = no caching
}
// Optional/auth paths controlled via skipPrefixes / skipExact / optionalExact.
```

### Use Case with UserRepository
```go
func (uc *MeUseCase) Execute(ctx context.Context, input MeInput) (*MeOutput, error) {
    user, err := uc.userRepo.GetByLogtoUserID(ctx, input.LogtoUserID)
    if errors.Is(err, errs.ErrNotFound) {
        return &MeOutput{Email: input.Email, NeedsSetup: true}, nil
    }
    // ... return full profile with NeedsSetup=false
}
```

## Naming Conventions
| Type | Convention | Example |
|------|-----------|---------|
| Files | snake_case | `create_card_brand.go` |
| Test Files | `*_test.go` + same/external pkg | `middleware_test.go`, `register_test.go` |
| Generated Mocks | `mocks.gen.go` per package | `mocks.gen.go` |
| Packages | lowercase | `cardbrand`, `auth`, `usecases` |
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
- `GOEXPERIMENT=jsonv2` required for all `go build`/`go test`/`go run` invocations (jwx v4)
- Extract interfaces for testability (see `jwkFetcher` in auth middleware)
- Graceful degradation: non-critical infra failures (cache down) → warn + fallback
- Integration tests use testcontainers-go (real Valkey container per test suite)
- Compile-time interface satisfaction checks: `var _ Interface = (*Concrete)(nil)`
- Use cases depend on interfaces (in ports/ or package-level), never concrete gateway types
- Use external test packages (`package pkg_test`) for use case unit tests with mockery mocks

## Security Requirements
- Validate all input via Huma schema tags
- Parameterized queries (sqlc generates)
- Environment-based secrets via `conf`
- SSL/TLS for DB (`PG_SSL_MODE=require`)
- No sensitive data in logs
- Proper error messages (don't leak internals)
- JWT validation via jwx/jwkfetch with Logto JWKS (15min cached TTL)
- Auth middleware: prefix skip for public paths, exact match for auth paths
- Bearer token extraction; optional paths populate claims but never reject
- `LOGTO_APP_CLIENT_ID` must match Logto API Resource identifier (aud claim)
- Logto M2M client credentials grant with token caching (60s safety buffer)
- JWKS TTL-based cache (15min) — cache errors degrade gracefully to direct fetch
- Unregistered users (JWT w/o DB record) receive 403 "needs setup"
- `/auth/me` uses optional auth — returns email for pre-fill when NeedsSetup=true

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
- **Auth Middleware**: `internal/gateways/http/v1/auth/middleware.go`
- **Auth Use Cases**: `internal/app/usecases/auth/` (register, me, errors)
- **Profile Use Cases**: `internal/app/usecases/profile/` (setup)
- **Logto Client**: `internal/gateways/logto/m2m_client.go`
- **Cache Client**: `pkg/cache/client.go`
- **Telemetry**: `pkg/telemetry/` (options, tracing, metrics, logging)
- **Migration**: `internal/gateways/postgres/migrations/`
- **SQL Queries**: `internal/gateways/postgres/sqlc/queries/`

## Related Context Files
- [golang-patterns.md](./golang-patterns.md) - Go patterns, testing, security
- [database-patterns.md](./database-patterns.md) - PostgreSQL, sqlc, migrations
- [docker-patterns.md](./docker-patterns.md) - Containerization patterns
- [workflow.md](./workflow.md) - Development workflow & quality
