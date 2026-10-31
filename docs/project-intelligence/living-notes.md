<!-- Context: project-intelligence/living | Priority: medium | Version: 1.6 | Updated: 2026-10-11 -->

# Living Notes

**Purpose**: Active issues, technical debt, open questions, and in-flight work. Review regularly; archive resolved items.

## In-Flight Work

| Item | Status | Where |
|------|--------|-------|
| `conf` → `koanf` migration | Done 2026-10-31 — executed | `docs/plans/conf-to-koanf-migration.md` |
| DOX AGENTS.md migration | Done 2026-08-12 — validate as used | Root `AGENTS.md` + 15 child docs |

## Technical Debt / Bugs Found (2026-08-12 audit)

- **No `go:generate` directives**: mock generation is Makefile-driven (`.mockery.yml`), zero `//go:generate` in source — fine, but note it
- **Stale docs risk**: previous `technical-domain.md` had drifted (Huma 2.38→2.39.1, jwkfetch module move, Go 1.26.3→1.26.4) — DOX tree is now source of truth; keep it current via DOX pass

## Known Constraints

- **golangci-lint cannot read Go 1.27.2 export data** (found 2026-10-11): Go 1.27.2 bumped the unified-IR export data format to V5 (`internal/pkgbits.V5`, go.dev/issue/81188). golangci-lint binaries up to v2.14.0 are built with go1.27.0 and support max V4, so linting a Go 1.27.2 toolchain fails with `could not load export data: internal error in importing "internal/goarch" (cannot decode "internal/goarch", export data version 5 is greater than maximum supported version 4)`. Mitigations: the CI Code Quality job runs on Go 1.27.1, and `make lint` runs inside the pinned golangci-lint image, which bundles Go 1.27.1. Revert the CI pin to `1.27` once a golangci-lint release built with Go >= 1.27.2 exists. Note `go.mod` must stay at `go 1.27` (never `go 1.27.x`), otherwise the lint image's `GOTOOLCHAIN=auto` would fetch a newer toolchain inside the container and trip the same failure.

## Open Questions

- Card/Bill/Transaction/Split: DB schema fully migrated (32 migrations: `person`, `card`, `bill`, `transaction`, `card_person`, `transaction_person` + FKs/indexes/triggers), but **app layer missing** (no entities, sqlc queries, ports, use cases, or handlers) — when do these modules get built?
- Which future features are prioritized? (business-domain.md Future Considerations: web app high, import formats high, payments/notifications/reports medium, multi-currency/mobile/alternating low)

## Resolved

- 2026-10-11: `gofrs/uuid/v5` → stdlib `uuid` — mockery v3.8.0 unblocked it (v3.7.3 couldn't parse go1.27 modules); the migration and the `GOEXPERIMENT=jsonv2` removal landed together (`docs/plans/uuid-stdlib-migration.md`)
- 2026-08-13: `last_modified_date` trigger — verified applied to all 6 entity tables (`user`, `card_brand`, `person`, `card`, `bill`, `transaction`); pivot tables (`card_person`, `transaction_person`) intentionally use `created_date` + `end_date` history (SPLIT-004 pattern) instead
- 2026-08-13: `device_poll.go` — tests existed but were co-located in `device_auth_test.go`; extracted to `device_poll_test.go` with fuller coverage (whitespace input, unexpected error pass-through)
- 2026-08-13: Profile package — `LogtoUserUpdater` extracted to `interfaces.go` (+ compile-time check); mockery entry added, manual `mockLogtoUpdater` replaced with `profile.NewMockLogtoUserUpdater(t)`
- 2026-08-13: M2M retry cap — code fixed `5s` → `10s` (`defaultRetryCap` in `m2m_client.go`); comment was correct
- 2026-08-13: `sqlc.yaml` — confirmed exists at repo root with correct overrides; no action needed
- 2026-08-12: Device token revocation shipped (#219) — endpoint + use case + docs
- 2026-08-12: Request ID propagation shipped (#231) — slog-context removed

## 📂 Codebase References

- **Plans**: `docs/plans/`
- **M2M client**: `internal/gateways/logto/m2m_client.go`
- **Auth use cases**: `internal/app/usecases/auth/`
- **Mock config**: `.mockery.yml`
