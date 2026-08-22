<!-- Context: internal/domain | Priority: critical | Version: 1.1 | Updated: 2026-08-22 -->

# internal/domain — Entities, domain errors, transactioner, shared value objects

**Purpose**: The innermost layer: pure business models (`entity/`), sentinel domain errors (`errs/`), the transaction abstraction (`Transactioner`), and shared domain value objects (`Pagination`). Must stay free of any dependency on application or infrastructure layers.

## Ownership

- `internal/domain/entity/` — `User`, `CardBrand`, `UserClaims` (+ tests)
- `internal/domain/errs/` — sentinel errors (`errs.go` + test)
- `internal/domain/transactioner.go` — `Transactioner` interface, `TransactionFunc`, `HasTX`/`WithTx` context markers
- `internal/domain/pagination.go` — `Pagination` value object (page-bound list options + `Offset()`); embed it in new list filter options instead of duplicating PageSize/PageNumber fields
- `internal/domain/mocks.gen.go` — generated `MockTransactioner`
- Must NOT import `internal/app/**`, `internal/gateways/**`, or `pkg/**` — only stdlib + `gofrs/uuid/v5`

## Local Contracts

- Entities are thin data holders: `gofrs/uuid/v5` IDs, `json` tags, timestamps; no business methods
- `User.LogtoUserID` carries `json:"-"` — never serialized or returned by the API (security)
- Domain errors are sentinel values (`errors.New`) grouped by concern; consumers check with `errors.Is()`, sources wrap with `%w`
- `Transactioner.WithTx(ctx, fn)` + `TransactionFunc` — the only way use cases run transactions
- `HasTX`/`WithTx` context helpers are the single source of "am I in a transaction?" — do not invent new context keys
- Implementers assert conformance: `var _ domain.Transactioner = (*TxManager)(nil)`
- Changing any exported interface/type here requires regenerating mocks (`make generate-mocks`)

## Work Guidance

Adding a sentinel error (pattern in `errs.go`):

```go
var (
    ErrCardBrandNotFound = errors.New("card brand not found")
)
```

Consumers compare with `errors.Is(err, errs.ErrCardBrandNotFound)` — never `err ==`. Entities: `uuid` ID + `CreatedDate`/`LastModifiedDate` (DB trigger fills the latter). Tests are table-driven (`transactioner_test.go` pattern).

## Verification

```bash
GOEXPERIMENT=jsonv2 go test -short ./internal/domain/...
make test          # unit tests
make code-check    # format + lint
```

## Child DOX Index

No child AGENTS.md files needed — `entity/` and `errs/` are owned by this doc.

## 📂 Codebase References

- `internal/domain/entity/user.go`
- `internal/domain/entity/card_brand.go`
- `internal/domain/entity/user_claims.go`
- `internal/domain/errs/errs.go`
- `internal/domain/transactioner.go`
- `internal/domain/pagination.go` (+ `pagination_test.go`)
- `internal/domain/mocks.gen.go`
