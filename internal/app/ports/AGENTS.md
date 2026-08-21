<!-- Context: internal/app/ports | Priority: critical | Version: 1.0 | Updated: 2026-08-12 -->

# internal/app/ports — Repository & service contracts

**Purpose**: Defines the interfaces (ports) that use cases depend on at the use case ⟷ gateway boundary. Contracts only — no implementations, no concrete gateway types.

## Ownership

- `internal/app/ports/user_repo.go` — `UserRepository`
- `internal/app/ports/card_brand_repo.go` — narrow card-brand role interfaces + option structs
- `internal/app/ports/claims.go` — `ClaimsProvider`
- `internal/app/ports/mocks.gen.go` — mockery-generated mocks for every interface here
- Must NOT: implement anything, import `internal/gateways/**` or `internal/app/usecases/**`, hold state

## Local Contracts

- Consumer-defined narrow interfaces, one method family each: `CreateCardBrandRepository`, `GetCardBrandByIDRepository`, `ListCardBrandRepository`, `UpdateCardBrandRepository`, `DeleteCardBrandRepository` (not one fat `CardBrandRepository`)
- Method signatures: `ctx` first, return `(*entity.X, error)`; input bundles go in option structs (`ListCardBrandFilterOptions`, `UpdateCardBrandOptions`)
- Doc comments state the error contract per method (e.g. "Returns ErrDuplicate if logto_user_id already exists")
- Imports limited to `context`, `entity`, `uuid`
- Implementers (gateways) assert satisfaction: `var _ ports.UserRepository = (*postgres.UserRepository)(nil)`
- No `go:generate` directives in source — mocks are driven by `.mockery.yml` (mockery v3.7.3 via Docker, testify template)

## Work Guidance

Adding a contract:

```go
// card_repo.go
type CreateCardRepository interface {
    Create(ctx context.Context, name string) (*entity.Card, error)
}
```

1. Define narrow interface + option structs + doc comments with error contracts
2. Regenerate mocks: `make generate-mocks` (or `make generate` for sqlc + mocks)
3. Implement in the gateway package with a compile-time satisfaction check
4. Keep interfaces small — a use case declares only what it calls

## Verification

```bash
make generate-mocks                              # regenerate mocks.gen.go after interface changes
GOEXPERIMENT=jsonv2 go test -short ./internal/app/...
make code-check                                  # format + lint (includes mocks.gen.go)
```

`make generate` (sqlc + mocks) also covers this package's mocks. Never hand-edit `mocks.gen.go`.

## Child DOX Index

No child AGENTS.md files needed — flat package. Sibling use cases have their own docs (`internal/app/usecases/*/AGENTS.md`).

## 📂 Codebase References

- `internal/app/ports/user_repo.go`
- `internal/app/ports/card_brand_repo.go`
- `internal/app/ports/claims.go`
- `internal/app/ports/mocks.gen.go`
- `.mockery.yml` — mock generation config
