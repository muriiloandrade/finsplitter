<!-- Context: project-intelligence/business-tech-bridge | Priority: high | Version: 1.2 | Updated: 2026-08-13 -->

# Business ↔ Technical Bridge

**Purpose**: Maps business needs and domain concepts to technical implementation — which module/AGENTS.md owns what, and how the business workflows are realized. Sources: `business-domain.md` (domain spec), the AGENTS.md DOX tree, and `technical-domain.md`.

## Business → Module Map

| Business concept | Technical home | Key files | AGENTS.md |
|------------------|----------------|-----------|-----------|
| User identity, auth lifecycle | `internal/app/usecases/auth/` | `register.go`, `device_*.go`, `me.go` | `internal/app/usecases/auth/AGENTS.md` |
| Identity via Logto (OIDC) | `internal/gateways/logto/` | `m2m_client.go`, `device_flow.go` | `internal/gateways/logto/AGENTS.md` |
| Card brand catalog | `internal/app/usecases/card-brand/` + `internal/gateways/http/v1/card-brand/` | `create_card_brand.go` etc. | `card-brand/AGENTS.md` + `http/v1/AGENTS.md` |
| User profile setup | `internal/app/usecases/profile/` | `setup.go` | `internal/app/usecases/profile/AGENTS.md` |
| User/CardBrand persistence | `internal/app/ports/` + `internal/gateways/postgres/` | `user_repo.go`, `card_brand_repo.go`, `sqlc/` | `ports/AGENTS.md` + `postgres/AGENTS.md` |
| Card, Bill, Transaction, Split (planned) | Schema migrated (32 migrations); app layer missing — future usecase/gateway modules | `internal/gateways/postgres/migrations/` | — |

## Workflow → Implementation

| Workflow (business) | How it maps technically |
|---------------------|------------------------|
| Manual transaction entry | Future: card/bill/transaction use cases + sqlc queries; auto bill assignment per BILL-004/006 |
| Transaction import (CSV/OFX) | Future: import use case + parser; no splits on bulk import |
| Recurring propagation | Future: bill-creation flow links recurring/installment transactions via `transaction_bill` pivot |
| Settlement calculation | Future: settlement use case aggregating transaction_person shares by pair |
| Auth (register, device flow, me) | **Implemented**: auth use cases → Logto gateway (M2M + device grant) → user repo |

## Key Architectural Contracts

- **Clean Architecture (Ports & Adapters)**: HTTP handlers (adapters) → use cases (core logic) → ports (interfaces) → postgres/logto (adapters). Business rules live in use cases + domain, never in gateways.
- **Consumer-defined interfaces**: use cases declare the gateway interfaces they need (e.g. `LogtoDeviceFlowClient` in auth); gateways satisfy them with compile-time checks (`var _ ... = (*logto.Client)(nil)`).
- **Auth strategy**: JWT+JWKS (standard tokens) + UserInfo fallback (opaque device-flow tokens); passwordless — device authorization grant only.
- **Request flow**: `internal/gateways/http/` (router + request ID) → `internal/gateways/http/v1/` (Huma handlers + middleware) → use cases.

## Gap: Business Rules vs Implementation

Business rules for cards, bills, transactions, and splits (`business-domain.md` §Business Rules: TXN-*, BILL-*, SPLIT-*, ALT-*) are **specified, and the DB schema is fully migrated** (32 migrations: `person`, `card`, `bill`, `transaction`, `card_person`, `transaction_person` + FKs/indexes/triggers) — but the **application layer is missing** (no entities, sqlc queries, ports, use cases, or handlers). Only auth, card-brand, and profile modules exist. When building these, follow the "Adding a New Feature" checklist in root `AGENTS.md`.

## 📂 Codebase References

- **Domain & business rules**: `business-domain.md`
- **Entry point / DI**: `cmd/api/main.go`
- **Routing**: `internal/gateways/http/v1/api.go`, `internal/gateways/http/v1/{auth,card-brand,profile}/routes.go`
- **Ports**: `internal/app/ports/`
- **Plans for future work**: `docs/plans/`
