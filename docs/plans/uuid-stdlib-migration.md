# Migration: `gofrs/uuid/v5` → stdlib `uuid` (Go 1.27)

> **Status**: Proposed
> **Target Go version**: 1.27 (stdlib `uuid` package is new in Go 1.27)
> **Dependencies removed**: `github.com/gofrs/uuid/v5`, `github.com/jackc/pgx-gofrs-uuid`, `github.com/google/uuid` (indirect)

## 1. Why

Go 1.27 ships a first-party [`uuid`](https://pkg.go.dev/uuid) package that generates, parses, and marshals UUIDs. Finsplitter uses `gofrs/uuid/v5` in ~16 production files and ~20 test files, plus a pgx adapter (`pgx-gofrs-uuid`) and a sqlc override. Adopting the stdlib package removes 2 direct dependencies and 1 indirect dependency, and aligns with the project's "prefer stdlib" direction.

## 2. Target API (stdlib `uuid`)

```go
type UUID [16]uint8

func New() UUID            // random UUID (v4)
func NewV4() UUID          // explicit v4
func NewV7() UUID          // v7 (timestamp-based)
func Parse(s string) (UUID, error)
func MustParse(s string) UUID
func Nil() UUID
func Max() UUID

func (u UUID) String() string
func (u UUID) MarshalText() ([]byte, error)
func (u UUID) UnmarshalText([]byte) error
func (u UUID) AppendText([]byte) ([]byte, error)
func (u UUID) Compare(UUID) int
```

**Key differences vs `gofrs/uuid`:**

| gofrs/uuid | stdlib uuid | Notes |
|---|---|---|
| `uuid.Must(uuid.NewV4())` | `uuid.New()` | stdlib `New()` returns `UUID` directly (no error) — no `Must` needed |
| `id.IsNil()` | `id == uuid.Nil()` | stdlib has no `IsNil()` method |
| `uuid.UUID` = `[16]byte` | `uuid.UUID` = `[16]uint8` | Same underlying type, interchangeable |
| `uuid.FromString(s)` | `uuid.Parse(s)` | Different name, same semantics |
| `uuid.Must(uuid.FromString(s))` | `uuid.MustParse(s)` | |

## 3. Changes needed

### 3.1 Go toolchain

- Bump `go 1.26.4` → `go 1.27` in `go.mod`.
- Verify local toolchain: `go version` must be ≥ 1.27.

### 3.2 sqlc configuration (`sqlc.yaml`)

Replace the two `gofrs/uuid` overrides with the stdlib type:

```yaml
overrides:
  - go_type:
      import: "uuid"
      type: "UUID"
    db_type: "uuid"
  - go_type:
      import: "uuid"
      type: "UUID"
    db_type: "uuid"
    nullable: true
```

> ⚠️ **Risk**: sqlc v1.30.0 may not resolve the stdlib `uuid` package (it has built-in support for `github.com/google/uuid` and `github.com/gofrs/uuid`). Verify with `make generate-sqlc`. If unsupported, either:
> - upgrade sqlc to a version that supports stdlib `uuid`, or
> - keep the override pointing at `github.com/google/uuid` (already an indirect dep) and use that type instead.

### 3.3 pgx integration (`internal/gateways/postgres/pool.go`)

- Remove `pgxuuid "github.com/jackc/pgx-gofrs-uuid"` import and the `pgxuuid.Register(conn.TypeMap())` call in `AfterConnect`.
- pgx v5 has native `pgtype.UUID` support. Verify that scanning a `uuid` column into `[16]byte`/`uuid.UUID` works out of the box (it should — pgx maps `uuid` → `pgtype.UUID` → `[16]byte`).
- If pgx cannot scan directly into the stdlib type, register a small custom codec in `AfterConnect` (see Appendix A).

### 3.4 Production code (16 files)

Mechanical import swap `github.com/gofrs/uuid/v5` → `uuid` in:

- `internal/domain/entity/user.go`
- `internal/domain/entity/card_brand.go`
- `internal/app/ports/card_brand_repo.go`
- `internal/app/ports/user_repo.go`
- `internal/app/usecases/card-brand/interfaces.go`
- `internal/app/usecases/card-brand/delete_card_brand.go`
- `internal/app/usecases/card-brand/get_card_brand_by_id.go`
- `internal/gateways/postgres/user.go`
- `internal/gateways/postgres/card_brand.go`
- `internal/gateways/postgres/sqlc/models.go` (regenerated)
- `internal/gateways/postgres/sqlc/card_brand.sql.go` (regenerated)
- `internal/gateways/postgres/sqlc/user.sql.go` (regenerated)
- `internal/gateways/http/v1/card-brand/delete_card_brand.go`
- `internal/gateways/http/v1/card-brand/list_card_brand.go`
- `internal/gateways/http/v1/card-brand/update_card_brand.go`
- `internal/gateways/http/v1/card-brand/get_card_brand.go`

**Non-mechanical changes:**

- `internal/gateways/postgres/card_brand.go:116` — `opts.ID.IsNil()` → `opts.ID == uuid.Nil()`.

### 3.5 Test code (~20 files)

Swap imports in all `*_test.go` files listed in the grep (auth, card-brand, profile, postgres, http v1). Replace:

- `uuid.Must(uuid.NewV4())` → `uuid.New()`
- Any `uuid.FromString(...)` → `uuid.Parse(...)` / `uuid.MustParse(...)`

### 3.6 go.mod cleanup

```bash
go mod tidy
```

Removes `gofrs/uuid/v5`, `pgx-gofrs-uuid`, and (if unused elsewhere) `google/uuid`.

## 4. Possible drawbacks / risks

| # | Risk | Severity | Mitigation |
|---|------|----------|------------|
| 1 | **sqlc may not support stdlib `uuid`** in v1.30.0 | High | Verify with `make generate-sql`; upgrade sqlc or fall back to `google/uuid` override |
| 2 | **pgx scanning** — `pgx-gofrs-uuid` adapter removed; native pgx UUID support must cover the stdlib type | Medium | Test `GetCardBrandByID` / `GetByID` against real DB; add custom codec if needed (Appendix A) |
| 3 | **Go 1.27 requirement** — stdlib `uuid` doesn't exist before 1.27 | Medium | Requires toolchain bump; CI/Dockerfile must use Go ≥ 1.27 |
| 4 | **API differences** — `IsNil()`, `FromString`, `Must` semantics differ | Low | Mechanical, covered by compiler + tests |
| 5 | **JSON marshaling** — stdlib `uuid` marshals via `MarshalText` (string form) — same as gofrs | Low | Verify one handler test |
| 6 | **Mock regeneration** — `mocks.gen.go` files reference `gofrs/uuid` | Low | Run `make generate-mocks` |

## 5. Execution order

1. Bump Go toolchain + `go.mod` to 1.27.
2. Update `sqlc.yaml` overrides; run `make generate-sql`; confirm generated code compiles with stdlib `uuid`.
3. Update `pool.go` (drop `pgx-gofrs-uuid`).
4. Swap imports in production files; fix `IsNil()`.
5. Swap imports in test files; fix `NewV4`/`FromString`.
6. `make generate-mocks`, `go mod tidy`.
7. `make test` + `make code-check`.

## Appendix A — custom pgx codec (only if native scan fails)

```go
// registerUUID registers the stdlib uuid.UUID type with pgx.
func registerUUID(tm *pgtype.Map) {
    tm.RegisterType(&pgtype.Type{
        Name:  "uuid",
        OID:   pgtype.UUIDOID,
        Codec: &pgtype.UUIDCodec{},
    })
}
```

(Adjust to the pgx v5 API in use; native support is expected, so this is a fallback.)