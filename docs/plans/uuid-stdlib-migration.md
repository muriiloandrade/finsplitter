# Migration: `gofrs/uuid/v5` → stdlib `uuid` (Go 1.27)

> **Status**: ⛔ **BLOCKED** — awaiting mockery Go 1.27 support (see §0)
> **Target Go version**: 1.27 (stdlib `uuid` package is new in Go 1.27)
> **Dependencies removed**: `github.com/gofrs/uuid/v5`, `github.com/jackc/pgx-gofrs-uuid`, `github.com/google/uuid` (indirect)

## 0. Blocker — mockery cannot parse Go 1.27 modules (discovered 2026-08-22)

The migration was executed through the code-swap stage on `feat/uuid-stdlib-migration` and reverted. Everything worked **except mock generation**:

| Check | Result |
|-------|--------|
| `go.mod` bump → 1.27 + full build | ✅ |
| sqlc 1.31.1 accepts stdlib `uuid` override + regenerates | ✅ (plan risk #1 cleared) |
| pgx native scan/encode after dropping `pgx-gofrs-uuid` | ✅ **runtime-verified** — full postgres integration suite green vs real Postgres 18.4 (testcontainers), no codec registration needed |
| sqlc-generated queries at runtime (create/get/update/delete/list/exists, both tables) | ✅ integration suite green |
| jwx-dependent packages flag-less (logto, auth UCs, http auth) | ✅ validation-gate step 1 green |
| `GOEXPERIMENT=jsonv2` inert on 1.27 (§3.7) | ✅ empirically verified |
| `make generate-mocks` (`vektra/mockery:v3.7.3` docker image) | ❌ image built with go1.26 → `package requires newer Go version go1.27 (application built with go1.26)` |
| Workaround: `go run github.com/vektra/mockery/v3@v3.7.3` under local go1.27 | ❌ `internal error: package "context" without types was imported from ".../internal/domain"` — v3.7.3's embedded x/tools loader predates go1.27 export data |

As of 2026-08-22, **v3.7.3 is the newest mockery release** (2026-08-12) — there is no newer version to pin.

### Validation scope (disposable worktree, 2026-08-22)

The sqlc/pgx layer was re-validated **in isolation** on a throwaway worktree (`validate/sqlc-stdlib-uuid`, removed afterwards): toolchain bump + sqlc overrides + entity/port/repo/test swaps, then `go test ./internal/gateways/postgres/` against a real Postgres 18.4 container — all green (~20s). Mock-dependent `*_repo_unit_test.go` files were excluded (mockery blocked). Usecase/handler layers were NOT part of this validation; they were already proven compile-level in the first migration attempt and only need the mechanical swap.

### Unblock criteria

A mockery release that both (a) parses go1.27 modules and (b) ships as a Docker image built with go ≥ 1.27. Renovate will bump the pinned image automatically once published.

### Resuming this migration

1. Wait for renovate's `vektra/mockery` bump PR (or check releases manually).
2. **Pre-flight before any code change**: bump `go.mod` only, then run `make generate-mocks`. If it fails, the blocker persists — stop.
3. Only then execute §5 in order; the §3.7 flag-less validation gates remain mandatory.

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

> ⚠️ **`Nil` is a function in stdlib** (`func Nil() UUID`), not a var like gofrs. Bare `uuid.Nil` references (comparisons, returns) will NOT compile — sed for `uuid.Nil` → `uuid.Nil()`, but review each site: `uuid.Nil(` must not become `uuid.Nil()(`. Found 4 bare sites during validation (test files + entity test).

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

> ✅ **Verified 2026-08-22**: sqlc 1.31.1 accepts this override and generates `uuid "uuid"` imports correctly (the plan's original risk referenced sqlc v1.30.0; the repo pins 1.31.1 via Makefile).

### 3.3 pgx integration (`internal/gateways/postgres/pool.go`)

- Remove `pgxuuid "github.com/jackc/pgx-gofrs-uuid"` import and the `pgxuuid.Register(conn.TypeMap())` call in `AfterConnect`.
- **No replacement codec or `RegisterDefaultPgType` call is needed.** pgx v5.10 encodes AND scans stdlib `uuid.UUID` natively (any `[16]byte` type — confirmed by jackc in [pgx#2636](https://github.com/jackc/pgx/issues/2636); `pgtype.UUID` itself can't alias the stdlib type until pgx's minimum Go is 1.27, but that doesn't affect us).
- Runtime proof: full postgres integration suite passes against real Postgres 18.4 with zero codec registration — notably `testutils/postgres.go` builds its pool via plain `pgxpool.ParseConfig` with no `AfterConnect` hook at all, so the tests exercise the unregistered path.
- Appendix A is retained only as a historical fallback; it was never needed.

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

- Same file — `opts.ID.IsNil()` → `opts.ID == uuid.Nil()`.

> Note: `ListCardBrands` used to need a special squirrel workaround here (§3.8 in earlier revisions). Squirrel was removed from the codebase on 2026-08-22 (`refactor: replace squirrel dynamic query with sqlc list card brands query`), so this step no longer exists — the method is pure sqlc now.

### 3.5 Test code (~20 files)

Swap imports in all `*_test.go` files listed in the grep (auth, card-brand, profile, postgres, http v1). Replace:

- `uuid.Must(uuid.NewV4())` → `uuid.New()`
- Any `uuid.FromString(...)` → `uuid.Parse(...)` / `uuid.MustParse(...)`

### 3.6 go.mod cleanup

```bash
go mod tidy
```

Removes `gofrs/uuid/v5`, `pgx-gofrs-uuid`, and (if unused elsewhere) `google/uuid`.

### 3.7 Remove `GOEXPERIMENT=jsonv2` (Go 1.27) — validation-gated

**Why it existed**: jwx v4 imports stdlib `encoding/json/v2` unconditionally; before Go 1.27 that package only existed under `GOEXPERIMENT=jsonv2`. Go 1.27 stabilized `encoding/json/v2` + `encoding/json/jsontext` (`$GOROOT/api/go1.27.txt`, proposal #71497), making the flag inert — verified empirically: identical marshal/unmarshal output with and without it on go1.27.0.

#### ⛔ Validation gate — MUST run BEFORE removing the flag from any file

Run the exact commands the Makefile runs, but with the flag unset. jwx-dependent packages first, then full suites:

```bash
# 1. jwx-priority packages (logto client → pkg/httpclient, auth use cases, auth handlers)
env -u GOEXPERIMENT go test -short ./internal/gateways/logto/... \
  ./internal/app/usecases/auth/... ./internal/gateways/http/v1/auth/...

# 2. Full unit suite (mirror of `make test`)
env -u GOEXPERIMENT go test -short ./...

# 3. Integration suite (mirror of `make test-int`)
env -u GOEXPERIMENT go test ./...

# 4. E2E (mirror of `make test-e2e`)
env -u GOEXPERIMENT go test -tags=e2e -v -run "^TestE2E" \
  ./internal/gateways/http/v1/auth/...
```

**Gate rule**: all four green → proceed to stripping below. Any failure attributable to JSON behavior differences → STOP, keep the flag, report before changing anything.

> Note: steps 2–4 only become runnable after §3.4–3.6 complete (mid-migration uuid type mismatches break unrelated builds). Step 1 can run as soon as Batch 1 lands.

#### Files to strip (only after the gate passes)

Functional:

| File | Sites |
|------|-------|
| `Makefile` | lines 44 (`GOLANGCI_LINT_CMD`), 126 (`test`), 130 (`test-int`), 134 (`test-e2e`), 142 (`test-cov`) |
| `Dockerfile` | line 6 (`ENV`), line 21 (`RUN go build`) |
| `compose.yml` | line 20 (environment) |
| `.github/workflows/test.yml` | line 39 |
| `.github/workflows/prerequisites.yml` | lines 35, 39 |
| `.github/workflows/quality.yml` | line 48 (env block) |
| `.github/workflows/security.yml` | lines 27, 54 |

Docs / DOX pass (same commit or follow-up):

- Root `AGENTS.md` (×2), `cmd/api/AGENTS.md` (×2), `pkg/{telemetry,logctx,httpclient,cache}/AGENTS.md`, `internal/{domain,config}/AGENTS.md`, `internal/app/ports/AGENTS.md`
- `docs/project-intelligence/{decisions-log,technical-domain,navigation}.md`
- `.env.example` (comment block)

## 4. Possible drawbacks / risks

| # | Risk | Severity | Mitigation |
|---|------|----------|------------|
| 1 | ~~**sqlc may not support stdlib `uuid`**~~ | ~~High~~ → **cleared** | sqlc 1.31.1 verified working (§3.2) |
| 2 | ~~**pgx scanning**~~ | ~~Medium~~ → **cleared** | Runtime-verified vs real PG 18.4; no codec needed (§3.3) |
| 3 | **Go 1.27 requirement** — stdlib `uuid` doesn't exist before 1.27 | Medium | Requires toolchain bump; CI/Dockerfile must use Go ≥ 1.27 |
| 4 | **API differences** — `IsNil()`, `FromString`, `Must` semantics differ; `Nil` is a function | Low | Mechanical, covered by compiler + tests (§2 note) |
| 5 | **JSON marshaling** — stdlib `uuid` marshals via `MarshalText` (string form) — same as gofrs | Low | Verify one handler test |
| 6 | **Mock regeneration** — `mocks.gen.go` files reference `gofrs/uuid` | **Blocker** | ⛔ mockery has no Go 1.27 support yet (§0) |

> Former risk #7 (squirrel expanding `[16]byte` into IN-lists in dynamic SQL) was removed on 2026-08-22: squirrel is no longer a dependency, so the risk class is gone.

## 5. Execution order

0. **Pre-flight (blocker check)**: bump `go.mod` to 1.27 only, run `make generate-mocks`; if it fails, stop — see §0.
1. Bump Go toolchain + `go.mod` to 1.27.
2. Update `sqlc.yaml` overrides; run `make generate-sqlc`; confirm generated code compiles with stdlib `uuid`.
3. Update `pool.go` (drop `pgx-gofrs-uuid`; add nothing in its place — §3.3).
4. Swap imports in production files; fix `IsNil()`.
5. Swap imports in test files; fix `NewV4`/`FromString`/bare `uuid.Nil` (§2 note).
6. `make generate-mocks`, `go mod tidy`.
7. **Integration check**: `env -u GOEXPERIMENT go test ./internal/gateways/postgres/` (docker required) — catches runtime encoding failures that compile checks miss.
8. **Validation gate for `GOEXPERIMENT` removal** — run §3.7's flag-less commands (jwx packages → unit → integration → e2e); STOP on any failure.
9. Strip `GOEXPERIMENT=jsonv2` from Makefile, Dockerfile, compose.yml, workflows, `.env.example` (§3.7 table).
10. `make test` + `make code-check` (now flag-less by construction).
11. DOX pass: update AGENTS.md chain + project-intelligence docs (uuid mentions + GOEXPERIMENT mentions).

## Appendix A — custom pgx codec (verified NOT needed)

Kept for historical reference only. During the 2026-08-22 validation, pgx v5.10 encoded and scanned stdlib `uuid.UUID` natively with zero registration — the full postgres integration suite passed against real Postgres 18.4 without any codec, and the test harness never registers one. Do not add this unless a future pgx regression appears.