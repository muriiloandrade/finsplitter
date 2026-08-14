<!-- Context: cmd/api | Priority: critical | Version: 1.0 | Updated: 2026-08-12 -->

# cmd/api — Application entry point & dependency wiring

**Purpose**: Single entry point (`main.go`) for the finsplitter API. Owns startup order, dependency injection, and server lifecycle. No business logic — everything is composed from package constructors.

## Ownership

- `cmd/api/main.go` — the only file in the package
- Responsible for: config load → OTel init → DB pool + migrations → clients/repos → router → HTTP server start/graceful stop
- Must NOT contain business logic, SQL, handler implementation, or config parsing rules
- Panic on fatal startup failure; graceful degradation only where designed (Redis)

## Local Contracts

- All wiring lives in `main.go` via small `newXxx` constructors (e.g. `newCardBrandAPI`, `newAuthAPI`, `newRedisClient`)
- Build info (`BuildCommit`, `BuildTag`, `BuildTime`) injected via ldflags — never hardcoded
- Startup order is fixed: config → telemetry → DB pool → migrations → clients → router → server
- Redis failure = warn + `nil` client (auth middleware then fetches JWKS per request) — non-fatal
- Shutdown: 5s graceful timeout (`gracefulShutdownTimeout`), then close pool/redis/otel
- Compile/build with `GOEXPERIMENT=jsonv2` (jwx v4 dependency)

## Work Guidance

Wiring a new feature module (card-brand pattern):

```go
func newCardBrandAPI(pgTxManager *postgres.TxManager) cbHandler.API {
    repo := postgres.NewCardBrandRepository(pgTxManager)
    uc := cbUCs.NewCreateCardBrandUC(repo, pgTxManager)
    return cbHandler.API{
        CreateCardBrandHandler: cbHandler.NewCreateCardBrandHandler(&uc).CreateCardBrand,
    }
}

// in main(): add to the v1.API struct, then apiV1.Routes(router)
apiV1 := v1.API{ /* existing fields */ CardBrandAPI: newCardBrandAPI(pgTxManager) }
```

Order matters: create repos → use cases → handler APIs → register routes → start server.

## Verification

```bash
GOEXPERIMENT=jsonv2 go build ./cmd/api/...   # compile check
make test                                     # unit tests (short mode)
make code-check                               # format + lint
```

No `_test.go` in this package — behavior is covered by tests in the wired packages.

## Child DOX Index

No child AGENTS.md files needed — the package contains only `main.go`. Everything it wires is owned by its own doc (`internal/config`, `internal/gateways/http/v1`, `internal/gateways/postgres`, `pkg/telemetry`, etc.).

## 📂 Codebase References

- `cmd/api/main.go` — entry point, DI, lifecycle
- `internal/gateways/http/request_id.go` — router middleware wired via `_http.NewRouter`
- `internal/gateways/http/v1/` — route registration (`apiV1.Routes`)
- `internal/gateways/postgres/migrations` — startup migrations
- `pkg/telemetry/` — OTel providers via functional options
