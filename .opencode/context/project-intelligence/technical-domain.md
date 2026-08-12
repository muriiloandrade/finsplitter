<!-- Context: project-intelligence/technical | Priority: critical | Version: 1.7 | Updated: 2026-07-04 -->

# Technical Domain

> Technical foundation, architecture, and development patterns for Finsplitter — a financial expense management and splitting system.

## Quick Reference

- **Purpose**: Understand how Finsplitter works technically
- **Stack**: Go backend API, PostgreSQL, Clean Architecture
- **Entry**: `cmd/api/main.go`
- **Update Triggers**: Tech stack changes | New patterns | Architecture decisions

---

## Primary Stack

| Layer | Technology | Version | Rationale |
|-------|-----------|---------|-----------|
| Language | Go | 1.26.3 | Type safety, performance, single binary |
| API Framework | Huma v2 | 2.38.0 | OpenAPI-first REST, generates spec from code |
| Database | PostgreSQL | 18 | Relational integrity, JSONB for flexible data |
| DB Driver | pgx v5 | 5.10.0 | Native PostgreSQL driver, connection pooling |
| Query Builder | sqlc | 1.30.0 | Type-safe SQL generation, no ORM overhead |
| Migrations | golang-migrate | 4.19.1 | Versioned SQL migrations |
| Cache | Valkey + go-redis/v9 | 9.21.0 | Redis-compatible JWKS cache, rate limiting |
| Auth | Logto | latest | OIDC/OAuth2 SSO, self-hosted |
| JWT | lestrrat-go/jwx v4 + jwkfetch/v4 | 4.0.2 | Logto-compatible JWT validation, JWKS fetching |
| Observability | OpenTelemetry | 1.44.0 | Traces, metrics, logs (OTLP exporter) |
| Config | ardanlabs/conf | — | Environment-based configuration |
| Logging | stdlib log/slog + pkg/logctx | — | Context-aware structured logging |
| Mock Generation | mockery v3 | 3.7.1 | Interface mock generation (Docker) |
| Integration Testing | testcontainers-go | 0.43.0 | Real containers per test suite |
| HTTP Client | resty v3 | 3.0.0-rc.2 | REST client for Logto Management API |
| Money Math | shopspring/decimal | 1.4.0 | High-precision decimal arithmetic |
| Linting | golangci-lint | 2.11.3 | Format + lint (Docker) |
| Build | GOEXPERIMENT=jsonv2 | — | Required (jwx v4 depends on encoding/json/v2) |

---

## Architecture

**Pattern**: Clean Architecture (Ports & Adapters)
**Type**: Modular Monolith

```
cmd/api/main.go              # Entry point, DI wiring
internal/
├── config/                  # ardanlabs/conf-based env loading
├── domain/
│   ├── entity/              # Business models (User, Card, Bill, Person, Transaction...)
│   └── errs/                # Domain error sentinels
├── app/
│   ├── ports/               # Repository interfaces (contracts)
│   └── usecases/            # Business logic (auth/, card-brand/, profile/)
└── gateways/
    ├── http/v1/             # Huma v2 HTTP handlers (auth/, card-brand/, profile/)
    ├── postgres/            # pgx + sqlc + migrations
    └── logto/               # Logto clients (M2M Management API + device flow)
pkg/
├── cache/                   # Valkey/Redis + OTel instrumentation
├── httpclient/              # Resty v3 wrapper
└── telemetry/               # OTel setup (traces, metrics, logs)
```

### Why This Architecture?

- **Testability**: Business logic is isolated from infrastructure (DB, HTTP)
- **Swapability**: PostgreSQL can be swapped without touching use cases
- **REST/OpenAPI-first**: Huma v2 generates OpenAPI spec from handler signatures
- **Single binary deploy**: Simple ops, Docker multi-stage build
- **Self-contained auth**: Logto runs alongside, handles OIDC/JWKS

---

## Key Technical Decisions

| Decision | Rationale | Impact |
|----------|-----------|--------|
| Clean Architecture | Isolate business logic from infrastructure | Easy to test and swap DB/HTTP |
| Huma v2 | OpenAPI spec generated from Go types | No separate spec maintenance |
| sqlc over ORM | Type-safe queries, no runtime overhead | Every query is reviewed SQL |
| pgx over database/sql | Native PostgreSQL, connection pooling | Better performance, pgx v5 pool |
| OpenTelemetry | Vendor-neutral observability | Swap exporters without code changes |
| Logto over Auth0/Firebase | Self-hosted, OIDC compliant | Full control, no vendor lock-in |
| Valkey over Redis | Drop-in Redis replacement, OSS | Same API, license-safe |
| Multi-stage Docker | Minimal production image (distroless) | Small attack surface, fast deploys |

---

## Project Structure

```
finsplitter/
├── cmd/api/main.go              # Entry point, DI wiring
├── internal/
│   ├── config/                  # Env config (ardanlabs/conf)
│   ├── domain/
│   │   ├── entity/              # Business models
│   │   └── errs/                # Domain error sentinels
│   ├── app/
│   │   ├── ports/               # Repository interfaces
│   │   └── usecases/            # Business logic (auth/, card-brand/, profile/)
│   └── gateways/
│       ├── http/v1/             # Huma v2 handlers + routes
│       ├── postgres/            # pgx, sqlc, migrations
│       └── logto/               # Logto M2M Management API client
├── pkg/
│   ├── cache/                   # Valkey/Redis + OTel instrumentation
│   ├── httpclient/              # Resty v3 wrapper
│   └── telemetry/               # OTel tracing, metrics, logs
├── scripts/                     # setup-m2m.sh, rotate-m2m-secret.sh
├── docs/
│   ├── ARCHITECTURE.md          # Full domain model & architecture
│   └── plans/                   # Implementation plans
├── Dockerfile                   # Multi-stage (setup → builder → production)
├── compose.yml                  # Backend profile
├── compose.infra.yml            # PostgreSQL + Valkey + Logto
└── compose.monitoring.yml       # OTel collector, Grafana, Tempo, Loki
```

---

## Integration Points

| System | Purpose | Protocol | Direction |
|--------|---------|----------|-----------|
| PostgreSQL | Persistent storage | pgx (TCP) | Internal |
| Valkey/Redis | Caching, sessions | RESP (TCP) | Internal |
| Logto | Authentication, OIDC | HTTP (REST/OIDC, device auth, UserInfo) | Internal |
| Grafana Tempo | Trace storage | OTLP (HTTP/gRPC) | Outbound |
| Grafana Loki | Log aggregation | OTLP (HTTP) | Outbound |
| Prometheus/Grafana | Metrics visualization | OTLP (HTTP) | Outbound |

---

## Technical Constraints

| Constraint | Origin | Impact |
|------------|--------|--------|
| jwx/v4 requires `GOEXPERIMENT=jsonv2` | Library dependency | Must set `GOEXPERIMENT=jsonv2` in env and IDE |
| Logto shares PostgreSQL | Infrastructure | Logto creates `logto` database, needs separate connection |
| Docker networking | Development | Services communicate via `finsplitter-net` bridge network |
| OIDC issuer mismatch | Docker vs host | `LOGTO_ISSUER` needs different value in compose vs IDE |

---

## Development Environment

```
Setup:         cp .env.example .env  &&  make start-infra
Requirements:  Go 1.26+, Docker, docker compose
Local Dev:     make start-dev          # Hot reload with compose watch
Debug Mode:    make start-debug        # Debug with delve
Testing:       make test               # go test ./...
Code Check:    make code-check         # Format + lint (golangci-lint)
Generate:      make generate           # sqlc + mocks
Migrations:    make new-migration name=create_table_foo
```

**Key Make Commands:**
- `make start-dev` — infra + hot reload
- `make new-migration name=create_table_foo` — timestamped migration
- `make generate` / `generate-sqlc` / `generate-mocks` — codegen
- `make test` / `make code-check` — test + format+lint

---

## Deployment

```
Environment:  Production / Development
Platform:     Docker (multi-stage build)
CI/CD:        Renovate (dep updates), Lefthook (git hooks)
Monitoring:   OTel → Grafana Tempo (traces) / Loki (logs) / Prometheus (metrics)
Image:        gcr.io/distroless/static-debian12:nonroot (production)
Build:        docker build --target production -t finsplitter:prod .
```

---

## Code Patterns

### API Handler (Huma v2)
```go
func (h CreateCardBrandHandler) CreateCardBrand(
    ctx context.Context, input *CreateCardBrandRequest,
) (*CreateCardBrandResponse, error) {
    logger := logctx.FromCtx(ctx)
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

### Auth Middleware (jwx + JWKS Cache + UserInfo Fallback)
```go
// Middleware validates tokens via two strategies:
// 1. JWT parsing + JWKS verification (for standard JWTs from M2M/SPA apps)
// 2. UserInfo fallback (for opaque tokens from device authorization flow)
// JWKS cached in Valkey with 15min TTL. Graceful degradation: cache failure → direct fetch.
// Interface extracted for testability (jwkFetcher → mockjwkFetcher).
type Middleware struct {
    jwkClient jwkFetcher       // *jwkfetch.Client in production
    cache     *cache.Client    // Valkey-backed; nil = no caching
    userInfoURL string         // Logto /oidc/me for opaque token validation
    httpClient  *http.Client   // HTTP client for UserInfo calls
}
// Public paths controlled via skipPrefixes / skipExact / optionalExact.
// Opaque token detection: if token has no dots (non-JWT), skips JWKS parse
// and goes directly to UserInfo endpoint.
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

---

### Device Flow — resty v3 Error Parsing

```go
// resty v3 uses SetResult for 2xx and SetResultError for non-2xx.
// Both must be set to capture error details on 4xx responses.
var result DeviceTokenResponse
resp, err := c.httpClient.R(ctx).
    SetFormData(formData).
    SetResult(&result).
    SetResultError(&result).  // Required: populates result.Error on 4xx
    Post(c.cfg.OIDCEndpoint + "/token")

if resp.IsStatusFailure() {
    switch result.Error {
    case "authorization_pending":
        return nil, ErrDeviceCodePending
    case "expired_token":
        return nil, ErrDeviceCodeExpired
    }
}
```

### Auth Handler Pattern — Concrete Use Case Types

```go
// The auth handler uses concrete *usecase.UseCase types directly — no local
// interfaces. Use cases already have their own interfaces for I/O (e.g.
// LogtoDeviceFlowClient, UserRepository); wrapping them again at the handler
// layer adds indirection without benefit.

type Handler struct {
    registerUC      *auth.RegisterUseCase
    meUC            *auth.MeUseCase
    deviceAuthUC    *auth.RequestDeviceAuthUseCase
    devicePollUC    *auth.PollDeviceTokenUseCase
    deviceRefreshUC *auth.RefreshDeviceTokenUseCase
}

func NewHandler(
    registerUC *auth.RegisterUseCase,
    meUC       *auth.MeUseCase,
    deviceAuthUC *auth.RequestDeviceAuthUseCase,
    devicePollUC *auth.PollDeviceTokenUseCase,
    deviceRefreshUC *auth.RefreshDeviceTokenUseCase,
) *Handler {
    return &Handler{
        registerUC:      registerUC,
        meUC:            meUC,
        deviceAuthUC:    deviceAuthUC,
        devicePollUC:    devicePollUC,
        deviceRefreshUC: deviceRefreshUC,
    }
}
```

### Device Flow — Public Handler Pattern

```go
// Device auth endpoints are fully public (no JWT required):
//   POST /auth/device           — initiate flow, returns device_code + user_code
//   POST /auth/device/poll      — poll for tokens after user approves in browser
//   POST /auth/device/refresh   — refresh tokens using a refresh token
// Registration is also public:
//   POST /auth/register         — passwordless, user then authenticates via device flow
//
// Middleware skips these via skipExact:
var skipExact = []string{
    "/auth/register",
    "/auth/device",
    "/auth/device/poll",
    "/auth/device/refresh",
}
```

### Consumer-Defined Interface Pattern

```go
// Interfaces are defined by the consumer (use case), not the producer (gateway).
// The use case declares what it needs; the concrete gateway satisfies it.
// Compile-time check in the gateway package ensures satisfaction.

// In use case package (consumer-owned):
type LogtoDeviceFlowClient interface {
    RequestDeviceCode(ctx context.Context) (*logto.DeviceCodeResponse, error)
    PollDeviceToken(ctx context.Context, deviceCode string) (*logto.DeviceTokenResponse, error)
    RefreshDeviceToken(ctx context.Context, refreshToken string) (*logto.DeviceTokenRefreshResponse, error)
}

// In gateway package — compile-time satisfaction check:
var _ auth.LogtoDeviceFlowClient = (*logto.Client)(nil)
```

> **Where this pattern applies**: use case ⟷ gateway boundary.  
> **Where it does NOT apply**: handler ⟷ use case boundary. Handlers use concrete `*usecase.UseCase` types directly. Extra interfaces at the handler layer add indirection without benefit (use cases already have their own interfaces for IO). See [Auth Handler Pattern — Concrete Use Case Types](#auth-handler-pattern--concrete-use-case-types) below.

---

## Naming Conventions

| Type | Convention | Example |
|------|-----------|---------|
| Files | snake_case | `create_card_brand.go` |
| Test Files | `*_test.go` + same/external pkg | `middleware_test.go` |
| Generated Mocks | `mocks.gen.go` per package | `mocks.gen.go` |
| Packages | lowercase | `cardbrand`, `auth`, `usecases` |
| Exported Functions | PascalCase | `GetCardBrandHandler` |
| Interfaces | PascalCase + suffix | `Repository`, `UseCase` |
| Database Tables | snake_case | `card_brands` |

---

## Code Standards

- Use `errors.Is()` for error checking (not equality)
- Domain errors in `internal/domain/errs/errs.go`
- Map DB errors via `pgerrcode.IsIntegrityConstraintViolation()`
- Use `logctx.FromCtx(ctx)` for structured logging
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

---

## Security Requirements

- Validate all input via Huma schema tags
- Parameterized queries (sqlc generates)
- Environment-based secrets via `conf`
- SSL/TLS for DB (`PG_SSL_MODE=require`)
- No sensitive data in logs
- Proper error messages (don't leak internals)
- Token validation via two strategies:
  1. **JWT+JWKS** (jwx/jwkfetch with Logto JWKS, 15min cached TTL) — for standard JWTs from M2M/SPA clients
  2. **UserInfo fallback** (Logto `/oidc/me` endpoint) — for opaque access tokens from device authorization flow
- Auth middleware: prefix skip for public paths, exact match for auth paths
- Bearer token extraction; optional paths populate claims but never reject
- Device flow endpoints (`/auth/device`, `/auth/device/poll`, `/auth/device/refresh`, `/auth/register`) are fully public (no auth)
- `LOGTO_APP_CLIENT_ID` must match Logto API Resource identifier (aud claim)
- Logto M2M client credentials grant with token caching (60s safety buffer)
- JWKS TTL-based cache (15min) — cache errors degrade gracefully to direct fetch
- Unregistered users (JWT w/o DB record) receive 403 "needs setup"
- `/auth/me` is optional auth; `/profile/setup` uses optional auth (handler verifies claims internally) — both return limited info when no token present
- `PATCH /profile/setup` is idempotent — returns 409 if user already fully registered
- Device flow uses Logto Native App client (`LOGTO_DEVICE_APP_CLIENT_ID`) separate from Traditional App client
- Password authentication disabled (must use device authorization grant)

---

## Adding New Feature (Step-by-Step)

1. `make new-migration name=create_table_foo`
2. Write SQL in `internal/gateways/postgres/sqlc/queries/foo.sql`
3. `make generate-sqlc` → generates Go code
4. Define entity in `internal/domain/entity/foo.go`
5. Define ports in `internal/app/ports/foo_repo.go`
6. Implement repo in `internal/gateways/postgres/foo.go`
7. Create use cases in `internal/app/usecases/foo/`
8. Create handlers in `internal/gateways/http/v1/foo/` using concrete `*usecase.UseCase` types (no local interfaces)
9. Wire in `cmd/api/main.go`
10. `make generate-mocks` + write tests

---

## 📂 Codebase References

- **Config**: `internal/config/config.go`
- **Domain Errors**: `internal/domain/errs/errs.go`
- **Auth Middleware**: `internal/gateways/http/v1/auth/middleware.go`
- **Auth Use Cases**: `internal/app/usecases/auth/` (register, device_auth, device_poll, device_refresh, me, errors)
- **Profile Use Cases**: `internal/app/usecases/profile/` (setup)
- **Logto Clients**: `internal/gateways/logto/m2m_client.go` (Management API), `internal/gateways/logto/device_flow.go` (device authorization grant)
- **Cache Client**: `pkg/cache/client.go`
- **Telemetry**: `pkg/telemetry/` (options, tracing, metrics, logging)
- **Migration**: `internal/gateways/postgres/migrations/`
- **SQL Queries**: `internal/gateways/postgres/sqlc/queries/`

---

## Related Context Files

- [`golang-patterns.md`](./golang-patterns.md) — Go patterns, testing, security
- [`database-patterns.md`](./database-patterns.md) — PostgreSQL, sqlc, migrations
- [`docker-patterns.md`](./docker-patterns.md) — Containerization patterns
- [`workflow.md`](./workflow.md) — Development workflow & quality
