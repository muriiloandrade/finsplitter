# AGENTS.md

> AI Agent Index - Project conventions and patterns for any AI coding assistant.

## Overview

Finsplitter is a Go backend API using Clean Architecture with Huma v2 and PostgreSQL.

**Quick Facts:**
- **Language**: Go 1.25
- **Framework**: Huma v2 (REST + OpenAPI)
- **Database**: PostgreSQL (pgx)
- **Architecture**: Clean Architecture (Ports & Adapters)

---

## Start Here

### Essential Commands
```bash
make start-dev      # Start infra + app with hot reload
make start-debug   # Start with delve debugging
make test          # Run tests
make code-check   # Format + lint
make generate     # Run sqlc + mocks generators
```

### First Time Setup
1. Copy `.env.example` to `.env` and configure
2. Run `make start-infra` to start database
3. Run `make start-dev` to start the app

---

## Project Structure

```
.
├── cmd/api/main.go           # Entry point, DI wiring
├── internal/
│   ├── config/               # Configuration (env vars)
│   ├── domain/
│   │   ├── entity/           # Business models
│   │   └── errs/             # Domain errors
│   ├── app/
│   │   ├── ports/            # Repository interfaces
│   │   └── usecases/         # Business logic
│   └── gateways/
│       ├── http/v1/          # HTTP handlers (Huma)
│       └── postgres/          # DB (pgx, sqlc)
└── pkg/telemetry/            # OTel tracing, metrics, logs
```

---

## Key Patterns

### Adding a New Entity (e.g., "Transaction")
1. `make new-migration name=create_table_transaction`
2. Write SQL in `internal/gateways/postgres/sqlc/queries/transaction.sql`
3. `make generate-sqlc`
4. Create entity: `internal/domain/entity/transaction.go`
5. Create port: `internal/app/ports/transaction_repo.go`
6. Implement repo: `internal/gateways/postgres/transaction.go`
7. Create use cases: `internal/app/usecases/transaction/`
8. Create handlers: `internal/gateways/http/v1/transaction/`
9. Wire in `cmd/api/main.go`
10. `make generate-mocks` + write tests

### Handler Pattern
```go
func (h Handler) HandlerName(ctx context.Context, input *Request) (*Response, error) {
    logger := slogctx.FromCtx(ctx)
    result, err := h.UseCase.Execute(ctx, input.Params)
    if errors.Is(err, errs.ErrDomainError) {
        return nil, huma.Error4xxYYY("message")
    }
    return &Response{Body: result}, nil
}
```

### Use Case with Transaction
```go
func (uc *UseCase) Execute(ctx context.Context, params string) (*Entity, error) {
    var result *Entity
    err := uc.tx.WithTx(ctx, func(ctx context.Context) error {
        r, err := uc.repo.Create(ctx, params)
        result = r
        return err
    })
    return result, err
}
```

---

## Naming Conventions

| Type | Convention | Example |
|------|-----------|---------|
| Files | snake_case | `create_card_brand.go` |
| Packages | lowercase | `cardbrand`, `usecases` |
| Exported | PascalCase | `GetCardBrandHandler` |
| Interfaces | PascalCase + suffix | `Repository`, `UseCase` |
| DB Tables | snake_case | `card_brands` |

---

## Code Standards

- Use `errors.Is()` for error checking
- Domain errors in `internal/domain/errs/errs.go`
- Map DB errors via `pgerrcode.IsIntegrityConstraintViolation()`
- Use `slogctx.FromCtx(ctx)` for structured logging
- Use sqlc for parameterized queries
- Configuration via `ardanlabs/conf`
- Test files `*_test.go` in same package

---

## Security

- Validate input via Huma schema tags
- Parameterized queries (sqlc generates)
- Environment-based secrets via `conf`
- SSL/TLS for DB (`PG_SSL_MODE=require`)
- No sensitive data in logs
- Use UUIDs for entity IDs

---

## Testing

```bash
# Run all tests
make test

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/app/usecases/...
```

Test pattern: Use testify/mock with table-driven tests in `*_test.go` files.

---

## Detailed Documentation

For complete patterns and examples, see:

- [`.opencode/context/project-intelligence/technical-domain.md`](.opencode/context/project-intelligence/technical-domain.md) - Core tech stack
- [`.opencode/context/project-intelligence/golang-patterns.md`](.opencode/context/project-intelligence/golang-patterns.md) - Go patterns
- [`.opencode/context/project-intelligence/database-patterns.md`](.opencode/context/project-intelligence/database-patterns.md) - PostgreSQL
- [`.opencode/context/project-intelligence/docker-patterns.md`](.opencode/context/project-intelligence/docker-patterns.md) - Docker
- [`.opencode/context/project-intelligence/workflow.md`](.opencode/context/project-intelligence/workflow.md) - Dev workflow
- [`.github/copilot-instructions.md`](.github/copilot-instructions.md) - Additional Copilot guidance

---

## Architecture Decision Records

See [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) for architectural decisions and rationale.
