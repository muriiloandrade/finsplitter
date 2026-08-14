<!-- Context: project-intelligence/living | Priority: medium | Version: 1.4 | Updated: 2026-08-13 -->

# Living Notes

**Purpose**: Active issues, technical debt, open questions, and in-flight work. Review regularly; archive resolved items.

## In-Flight Work

| Item | Status | Where |
|------|--------|-------|
| `conf` → `koanf` migration | Proposed — plan written | `docs/plans/conf-to-koanf-migration.md` |
| `gofrs/uuid/v5` → stdlib `uuid` | Proposed — target Go 1.27 | `docs/plans/uuid-stdlib-migration.md` |
| DOX AGENTS.md migration | Done 2026-08-12 — validate as used | Root `AGENTS.md` + 15 child docs |

## Technical Debt / Bugs Found (2026-08-12 audit)

- **No `go:generate` directives**: mock generation is Makefile-driven (`.mockery.yml`), zero `//go:generate` in source — fine, but note it
- **Stale docs risk**: previous `technical-domain.md` had drifted (Huma 2.38→2.39.1, jwkfetch module move, Go 1.26.3→1.26.4) — DOX tree is now source of truth; keep it current via DOX pass

## Open Questions

- Card/Bill/Transaction/Split: DB schema fully migrated (32 migrations: `person`, `card`, `bill`, `transaction`, `card_person`, `transaction_person` + FKs/indexes/triggers), but **app layer missing** (no entities, sqlc queries, ports, use cases, or handlers) — when do these modules get built?
- Which future features are prioritized? (business-domain.md Future Considerations: web app high, import formats high, payments/notifications/reports medium, multi-currency/mobile/alternating low)

## Resolved

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
