<!-- Context: gateways/http/v1 | Priority: critical | Version: 1.0 | Updated: 2026-08-12 -->

# HTTP v1 — Huma v2 API, handler groups, auth middleware

**Purpose**: Exposes the v1 REST API — OpenAPI spec generated from code via Huma v2, chi-level auth middleware, and thin handler groups (auth, card-brand, profile) that delegate to use cases.

## Ownership

- `api.go` — `API` struct (route-group aggregation), `Routes()`: auth middleware + health + single huma API + `RegisterRoutes`
- `health.go` — plain chi liveness/readiness handlers (no huma)
- `auth/` — `middleware.go`, `handler.go`, `routes.go`, `device.go` + tests (register, me, device flow)
- `card-brand/` — `routes.go` + one file per operation (`create_card_brand.go`, `get_card_brand.go`, `list_card_brand.go`, `update_card_brand.go`, `delete_card_brand.go`) + tests
- `profile/` — `handler.go`, `routes.go` + tests
- NOT: use cases, domain logic, Logto clients, OpenAPI config (`api/openapi.go`)

## Local Contracts

- **Handlers are THIN**: call use cases, map errors to huma errors. No business logic.
- Concrete `*usecase.UseCase` types — no handler-layer interfaces around use cases. The only local interface is `interfaceAPI` (route-group aggregation: `RegisterRoutes(api huma.API)`).
- **Huma v2 pattern**: request/response structs with `Body` + schema tags (`required`, `maxLength`, `example`, `doc`); handler signature `func(ctx context.Context, req *I) (*O, error)`; register via `huma.Register(api, huma.Operation{Method, Path, Summary, Description, Tags, Security, Errors}, handler)`.
- Validation is declarative — Huma schema tags validate input before the handler runs.
- Error mapping: `errors.Is(err, errs.X)` → `huma.Error4xx/5xx`. Never leak internals.
- `Security`: protected ops `[]map[string][]string{{"bearerAuth": {}}}`; optional ops `{{}, {"bearerAuth": {}}}`.
- Auth middleware is chi-level (applied in `Routes()` before route registration — chi requirement), not per-operation.
- Logging via `logctx.FromCtx(ctx)`.

## Auth Middleware (auth/middleware.go)

- `skipPrefixes`: `/health/`, `/docs`, `/openapi` — fully public.
- `skipExact` (public): `/auth/register`, `/auth/device`, `/auth/device/poll`, `/auth/device/refresh`, `/auth/device/revoke`.
- `optionalExact`: `/auth/me`, `/profile/setup` — populate claims when a valid token is present; no token → pass through; **invalid** token → 401.
- Token validation, two strategies:
  1. JWT via jwx + JWKS (Valkey cache, 15min TTL, key `jwks:keyset`); validates issuer + audience (`LOGTO_APP_CLIENT_ID`).
  2. UserInfo fallback (Logto `/oidc/me`) for opaque tokens only — a dotted token that fails JWKS is rejected, never sent to UserInfo.
- Cache down/miss → warn + direct fetch (graceful degradation). `cache == nil` → no caching.
- Protected paths: after validation, `userRepo.ExistsByLogtoUserID` — missing user → 403 needs setup.
- Claims via `GetUserClaims(ctx)` / `WithUserClaims(ctx)`; profile handler reads them through `ports.ClaimsProvider`.

## Work Guidance

1. New route group: create `v1/{module}/` with an `API` struct implementing `RegisterRoutes(api huma.API)`; `NewAPI` wires concrete use cases; register in the v1 `API` struct and `Routes()`.
2. New handler: structs with schema tags → handler method → expose as `HumaHandler` on the `API` → `huma.Register`.
3. New public/optional path: add to the skip lists in `middleware.go`, not in the route.
4. Middleware changes: keep the two-strategy validation and graceful degradation; update `middleware_test.go` / `middleware_userinfo_test.go`.

Example — thin handler delegating to a use case:

```go
func (h CreateCardBrandHandler) CreateCardBrand(ctx context.Context, input *CreateCardBrandRequest) (*CreateCardBrandResponse, error) {
    brand, err := h.UseCase.CreateCardBrand(ctx, input.Body.Name)
    if errors.Is(err, errs.ErrCardBrandAlreadyExists) {
        return nil, huma.Error409Conflict(err.Error())
    }
    return &CreateCardBrandResponse{*brand}, nil
}
```

## 📂 Codebase References

- `internal/gateways/http/v1/api.go`, `internal/gateways/http/v1/health.go`
- `internal/gateways/http/v1/auth/middleware.go`, `handler.go`, `routes.go`, `device.go`
- `internal/gateways/http/v1/card-brand/routes.go` + per-operation files
- `internal/gateways/http/v1/profile/handler.go`, `routes.go`
- `api/openapi.go` — OpenAPI config (`bearerAuth` security scheme)

## Verification

```bash
make test
make code-check
go test ./internal/gateways/http/v1/...
```

## Child DOX Index

- No child AGENTS.md files yet; v1 owns all handler groups. If a module grows its own conventions, its doc belongs at `internal/gateways/http/v1/{module}/AGENTS.md` (handler-pattern bucket from the DOX migration bundle).
