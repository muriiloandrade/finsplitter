# AGENTS.md

<!-- Context: dox/root | Priority: critical | Version: 2.2 | Updated: 2026-08-14 -->

# DOX framework

- DOX is a highly performant AGENTS.md hierarchy installed here
- Agent must follow DOX instructions across any edits

## Core Contract

- AGENTS.md files are binding work contracts for their subtrees
- Work products, source materials, instructions, records, assets, and durable docs must stay understandable from the nearest applicable AGENTS.md plus every parent AGENTS.md above it

## Read Before Editing

1. Read the root AGENTS.md
2. Identify every file or folder you expect to touch
3. Walk from the repository root to each target path
4. Read every AGENTS.md found along each route
5. If a parent AGENTS.md lists a child AGENTS.md whose scope contains the path, read that child and continue from there
6. Use the nearest AGENTS.md as the local contract and parent docs for repo-wide rules
7. If docs conflict, the closer doc controls local work details, but no child doc may weaken DOX

Do not rely on memory. Re-read the applicable DOX chain in the current session before editing.

## Update After Editing

Every meaningful change requires a DOX pass before the task is done.

Update the closest owning AGENTS.md when a change affects:

- purpose, scope, ownership, or responsibilities
- durable structure, contracts, workflows, or operating rules
- required inputs, outputs, permissions, constraints, side effects, or artifacts
- user preferences about behavior, communication, process, organization, or quality
- AGENTS.md creation, deletion, move, rename, or index contents

Update parent docs when parent-level structure, ownership, workflow, or child index changes. Remove stale or contradictory text immediately.

## Project Overview

Finsplitter is a Go backend API for expense management and splitting, using Clean Architecture (Ports & Adapters) with Huma v2 and PostgreSQL.

**Stack**: Go 1.26.4 · Huma v2.39.1 · PostgreSQL 18 (pgx v5) · sqlc · Valkey/go-redis v9 · Logto (OIDC) · jwx/v4 + jwkfetch · OpenTelemetry · resty v3 · ardanlabs/conf

**Build note**: `GOEXPERIMENT=jsonv2` required for all `go build`/`go test`/`go run` invocations (jwx v4 dependency).

## Essential Commands

```bash
make start-dev       # infra + app with hot reload
make test            # run tests
make code-check      # format + lint (golangci-lint)
make generate        # sqlc + mocks generators
make new-migration name=create_table_foo
```

## Commit Conventions

- Conventional commits (lowercase type: `feat`, `fix`, `docs`, `chore`, `refactor`, `test`, `ci`, ...) — enforced by lefthook locally and commitlint in CI
- **PR titles must pass the same commitlint rules** — the squash-merge commit is named after the PR title and is linted on push to `main`; CI also lints the title on pull requests
- Subject starts with a lowercase word (no sentence-case/uppercase acronyms at the start), header ≤ 100 chars, no trailing period
- Config: `commitlint.config.mjs` (repo root)

## First Time Setup

1. Copy `.env.example` to `.env` and configure
2. Run `make start-infra` to start database
3. Run `make start-dev` to start the app

## Naming Conventions

| Type | Convention | Example |
|------|-----------|---------|
| Files | snake_case | `create_card_brand.go` |
| Test Files | `*_test.go` + same/external pkg | `middleware_test.go` |
| Generated Mocks | `mocks.gen.go` per package | `mocks.gen.go` |
| Packages | lowercase | `cardbrand`, `usecases` |
| Exported Functions | PascalCase | `GetCardBrandHandler` |
| Interfaces | PascalCase + suffix | `Repository`, `UseCase` |
| DB Tables | snake_case | `card_brands` |

## Code Standards

- Use `errors.Is()` for error checking (not equality)
- Domain errors in `internal/domain/errs/errs.go`
- Map DB errors via `pgerrcode.IsIntegrityConstraintViolation()`
- Use `logctx.FromCtx(ctx)` for structured logging
- Use sqlc for parameterized queries (SQL-injection safe)
- Configuration via `ardanlabs/conf` with env vars
- Test files `*_test.go` with table-driven tests, testify/mock
- UUIDs for entity IDs (gofrs/uuid/v5), auto `last_modified_date` via trigger
- `GOEXPERIMENT=jsonv2` required for all go build/test/run
- Extract interfaces for testability
- Graceful degradation: non-critical infra failures (cache down) → warn + fallback
- Compile-time interface satisfaction checks: `var _ Interface = (*Concrete)(nil)`
- Use cases depend on interfaces (in ports/ or package-level), never concrete gateway types
- External test packages (`package pkg_test`) for use case unit tests with mockery mocks
- Handlers use concrete `*usecase.UseCase` types directly (no local interfaces at handler layer)
- Consumer-defined interfaces at the use case ⟷ gateway boundary

## Security Requirements

- Validate all input via Huma schema tags
- Parameterized queries (sqlc generates)
- Environment-based secrets via `conf`
- SSL/TLS for DB (`PG_SSL_MODE=require`)
- No sensitive data in logs; don't leak internals in error messages
- Token validation via two strategies:
  1. **JWT+JWKS** (jwx/jwkfetch, Logto JWKS, 15min cached TTL) — standard JWTs from M2M/SPA clients
  2. **UserInfo fallback** (Logto `/oidc/me`) — opaque tokens from device authorization flow
- Auth middleware: prefix skip for public paths, exact match for auth paths
- Public endpoints (no auth): `/auth/register`, `/auth/device`, `/auth/device/poll`, `/auth/device/refresh`
- Optional auth: `/auth/me`, `/profile/setup` (limited info when no token)
- `LOGTO_APP_CLIENT_ID` must match Logto API Resource identifier (aud claim)
- M2M client credentials grant with token caching (60s safety buffer)
- Device flow uses Logto Native App client (`LOGTO_DEVICE_APP_CLIENT_ID`)
- Password authentication disabled (device authorization grant only)

## Testing

```bash
make test                    # all tests
go test -cover ./...         # with coverage
go test ./internal/app/usecases/...   # specific package
```

Test pattern: testify/mock with table-driven tests in `*_test.go` files; integration tests via testcontainers-go.

## Adding a New Feature (Step-by-Step)

1. `make new-migration name=create_table_foo`
2. Write SQL in `internal/gateways/postgres/sqlc/queries/foo.sql`
3. `make generate-sqlc` → generates Go code
4. Define entity in `internal/domain/entity/foo.go`
5. Define ports in `internal/app/ports/foo_repo.go`
6. Implement repo in `internal/gateways/postgres/foo.go`
7. Create use cases in `internal/app/usecases/foo/`
8. Create handlers in `internal/gateways/http/v1/foo/` (concrete `*usecase.UseCase` types)
9. Wire in `cmd/api/main.go`
10. `make generate-mocks` + write tests

## Architecture Decision Records

See [`docs/project-intelligence/decisions-log.md`](docs/project-intelligence/decisions-log.md) for decisions and rationale; `business-domain.md` is the domain spec.

## Child DOX Index

| Path | Scope | Priority |
|------|-------|----------|
| `cmd/api/AGENTS.md` | Entry point, DI wiring | critical |
| `internal/config/AGENTS.md` | Env config (ardanlabs/conf) | high |
| `internal/domain/AGENTS.md` | Entities, domain errors, transactioner | critical |
| `internal/app/ports/AGENTS.md` | Repository interfaces (contracts) | critical |
| `internal/app/usecases/auth/AGENTS.md` | Register, device flow, me, refresh, revoke | critical |
| `internal/app/usecases/card-brand/AGENTS.md` | Card brand CRUD use cases | high |
| `internal/app/usecases/profile/AGENTS.md` | Profile setup use case | high |
| `internal/gateways/http/AGENTS.md` | Global router, request ID middleware | critical |
| `internal/gateways/http/v1/AGENTS.md` | Huma v2 handlers, routes, middleware | critical |
| `internal/gateways/logto/AGENTS.md` | Logto M2M + device flow clients | critical |
| `internal/gateways/postgres/AGENTS.md` | pgx, sqlc, migrations, testutils | critical |
| `pkg/cache/AGENTS.md` | Valkey/Redis client | high |
| `pkg/httpclient/AGENTS.md` | Resty v3 wrapper | high |
| `pkg/logctx/AGENTS.md` | Context-aware logging + request ID | high |
| `pkg/telemetry/AGENTS.md` | OTel tracing, metrics, logging | high |

Root-owned: `cmd/`, `internal/`, `pkg/` (routing via child docs above); `docs/`, `scripts/`, `Makefile`, compose files, Dockerfile, `.github/` (CI/CD).

## Docker & Deployment

- Multi-stage build: `setup` → `builder` → `production`; production image is `gcr.io/distroless/static-debian12:nonroot`
- Use specific base image tags (not `latest`); use `.dockerignore`; run as non-root; exec form for CMD/ENTRYPOINT; implement HEALTHCHECK
- Compose files: `compose.yml` (backend profile), `compose.infra.yml` (PostgreSQL + Valkey + Logto), `compose.monitoring.yml` (OTel, Grafana, Tempo, Loki)
- Security: scan images (`make docker-scout`), no secrets in images, read-only root FS + limited capabilities where possible

## Clean Code Principles

- **Constants**: no magic numbers — use named constants
- **Names**: descriptive, explain purpose (avoid unclear abbreviations)
- **Comments**: explain "why", not "what" (code should be self-documenting)
- **DRY**: extract repeated code into reusable functions
- **SRP**: each function does one thing

## Code Review Guidelines

- Verify information before presenting
- Review file-by-file changes
- No unnecessary confirmations
- Preserve existing code
- Single chunk edits

## Related Project Intelligence

The `docs/project-intelligence/` directory contains the knowledge index. `technical-domain.md` is now a lightweight index pointing to the DOX tree above.
