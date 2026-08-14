<!-- Context: app/usecases/card-brand | Priority: high | Version: 1.0 | Updated: 2026-08-12 -->

# Card Brand Use Cases — CRUD business logic

**Purpose**: Implements card brand CRUD (create, get, list, update, delete) as pure orchestration — validate input, run repo calls (writes inside a transaction), map errors to domain sentinels. No HTTP, no DB, no entity logic here.

## Ownership

- `interfaces.go` — use case contracts: `CreateCardBrandUseCase`, `GetCardBrandByIDUseCase`, `ListCardBrandsUseCase`, `UpdateCardBrandUseCase`, `DeleteCardBrandUseCase`
- `create_card_brand.go` — `CreateCardBrandUC` (write, transactional)
- `get_card_brand_by_id.go` — `GetCardBrandByIDUC` (read, direct repo call)
- `list_card_brand.go` — `ListCardBrandsUC` (read, direct repo call)
- `update_card_brand.go` — `UpdateCardBrandUC` (write, transactional)
- `delete_card_brand.go` — `DeleteCardBrandUC` (write, transactional)
- `mocks.gen.go` — mockery-generated mocks (repo + transactioner used by external tests)
- Tests — `*_test.go` per operation, `package usecases_test`
- NOT: HTTP handlers (`internal/gateways/http/v1/card-brand/`), Postgres repo impl, domain entity definitions

## Local Contracts

- **Package name is `usecases`** (directory is `card-brand/`). External tests import it aliased: `usecases "github.com/muriiloandrade/finsplitter/internal/app/usecases/card-brand"`.
- **Repo deps come from `internal/app/ports`**: `ports.CreateCardBrandRepository`, `GetCardBrandByIDRepository`, `ListCardBrandRepository`, `UpdateCardBrandRepository`, `DeleteCardBrandRepository`; options types `ports.ListCardBrandFilterOptions`, `ports.UpdateCardBrandOptions`.
- **Writes are transactional**: create/update/delete wrap the repo call in `domain.Transactioner.WithTx`; reads (get/list) call the repo directly — no `tx` field on read UCs.
- **Validation**: create — non-empty `name` (`errors.New("name is required")`); update — non-empty `name` + non-nil ID; delete — non-nil ID. Raw errors, not `errs` sentinels.
- **Error mapping via `errors.Is`**: create → `errs.ErrCardBrandAlreadyExists`; update → `errs.ErrCardBrandNotFound`, `errs.ErrCardBrandAlreadyExists`; get → `errs.ErrCardBrandNotFound`.
- **Handlers consume the concrete UC structs**; the `interfaces.go` contracts document what handler/DI expects — satisfied at compile time by the UC structs.
- Use cases never import concrete gateway/postgres types.

## Work Guidance

1. New read operation: struct with `ports.XRepository` dep → constructor → call repo → map `errs` via `errors.Is` → return entity.
2. New write operation: struct with `domain.Transactioner` + `ports.XRepository` → `tx.WithTx(ctx, func(ctx) error {...})` → map errors inside the closure → return the entity captured in an outer var.

Example — transactional create:

```go
err := uc.tx.WithTx(ctx, func(ctx context.Context) error {
    cb, err := uc.repo.CreateCardBrand(ctx, name)
    if errors.Is(err, errs.ErrCardBrandAlreadyExists) {
        return errs.ErrCardBrandAlreadyExists
    }
    insertedCardBrand = cb
    return err
})
```

3. Tests: table-driven, external package `usecases_test`, mock repo via `ports.NewMockXRepository(t)` + `domain.NewMockTransactioner(t)`.

## 📂 Codebase References

- `internal/app/usecases/card-brand/interfaces.go` — use case contracts
- `internal/app/usecases/card-brand/create_card_brand.go`, `get_card_brand_by_id.go`, `list_card_brand.go`, `update_card_brand.go`, `delete_card_brand.go`
- `internal/app/ports/card_brand_repo.go` — repository interfaces + option types
- `internal/domain/transactioner.go` — `Transactioner` / `TransactionFunc`
- `internal/domain/errs/errs.go` — card brand error sentinels

## Verification

```bash
make test
make code-check
go test ./internal/app/usecases/card-brand/...
```

## Child DOX Index

- No child AGENTS.md files needed — this is a leaf package. Root AGENTS.md owns global standards.
