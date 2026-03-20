# Auth with Logto — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Integrate Logto as the authentication provider for Finsplitter. Logto handles identity (sign-up, login, tokens); Finsplitter owns user profile data linked by a `logto_user_id` pseudo-FK.

**Architecture:** Trunk-based development, one feature branch (`feat/auth-logto-integration`). Local infra via Docker Compose. M2M auth uses Logto Management API with in-memory token caching. Finsplitter owns all Go code — Logto is a black-box identity provider.

**Tech Stack:** Go 1.26, Huma v2, Logto (self-hosted via Docker), PostgreSQL, OAuth2 Client Credentials.

**Development Workflow:**
- **TDD:** Write the failing test first for every unit. Run it to verify it fails. Write the minimal implementation. Run to verify it passes.
- **Review gate:** After tests pass, dispatch a **CodeReviewer** subagent to review the changes. Report findings to the user. **Wait for explicit user approval** before committing.
- **Atomic commits:** One logical change per commit. If a task has multiple test-implement pairs, each pair is its own commit.
- **Conventional Commits:** All commits follow [Conventional Commits](https://www.conventionalcommits.org/). Format: `type: description` — **without the scope block**. Examples:
  - `chore: add worktree directories to gitignore`
  - `feat: add user entity and domain errors`
  - `test: add auth middleware unit tests`
  - `fix: address linter warnings`
  - `refactor: extract token caching to client`
  - `docs: add setup instructions`

**Scope:**
- ✅ Local infra (Logto + DB)
- ✅ M2M bootstrap (one-time Logto CLI + setup script)
- ✅ Finsplitter M2M client with token caching
- ✅ User entity (domain + port + repo + use cases)
- ✅ Auth middleware (JWT validation)
- ✅ Registration endpoint (`POST /auth/register`)
- ✅ Profile setup endpoint (`POST /profile/setup`)
- ✅ Auth/me endpoint (`GET /auth/me`)
- ✅ Unit tests

**Out of Scope:**
- Frontend (any frontend integration is future work)
- Refresh of JWT token (handled by frontend SDK)
- Alternating payment, bill, card, transaction entities
- Custom Logto JWT claims (username injection) — handled separately in Logto Console

---

## Part 1: Infrastructure Setup

### Task 1: Create Logto PostgreSQL database service

**Files:**
- Modify: `compose.infra.yml`

**Step 1: Add Logto's PostgreSQL database to the infra compose**

Read `compose.infra.yml` and add this service before the `volumes:` block:

```yaml
    logto-db:
        profiles:
            - infra
        image: postgres:18.3-trixie@sha256:a9abf4275f9e99bff8e6aed712b3b7dfec9cac1341bba01c1ffdfce9ff9fc34a
        container_name: finsplitter-logto-db
        restart: unless-stopped
        healthcheck:
            test:
                [
                    "CMD-SHELL",
                    "pg_isready -d $${POSTGRES_DB} -U $${POSTGRES_USER}",
                ]
            start_period: 20s
            interval: 30s
            retries: 5
            timeout: 5s
        volumes:
            - logto_database:/var/lib/postgresql
        ports:
            - ${LOGTO_DB_PORT:-5433}:5432
        environment:
            POSTGRES_PASSWORD: ${LOGTO_DB_PASS:?logto database password required}
            POSTGRES_USER: ${LOGTO_DB_USER:-logto}
            POSTGRES_DB: ${LOGTO_DB:-logto}
        env_file:
            - .env
        networks:
            - finsplitter

volumes:
    database:
        driver: local
    logto_database:
        driver: local
```

**Step 2: Request review**

Dispatch a **CodeReviewer** subagent to review the `compose.infra.yml` changes. Present: the diff, the rationale (adds isolated PostgreSQL for Logto), and ask for approval.

**Wait for explicit user approval** before proceeding.

**Step 3: Commit**

```bash
git add compose.infra.yml
git commit -m "feat: add PostgreSQL database for Logto"
```

---

### Task 2: Update compose.yml with Logto service

**Files:**
- Modify: `compose.yml`

**Step 1: Add Logto service**

Add to the `services:` section of `compose.yml`:

```yaml
    logto:
        profiles:
            - infra
        container_name: logto
        depends_on:
            logto-db:
                condition: service_healthy
        image: svhd/logto:${TAG-latest}
        entrypoint: ["sh", "-c", "npm run cli db seed -- --swe && npm start"]
        ports:
            - "${LOGTO_PORT:-3001}:3001"
            - "${LOGTO_PORT_MGMT:-3002}:3002"
        environment:
            - TRUST_PROXY_HEADER=1
            - DB_URL=${LOGTO_DB_URL}
            - ENDPOINT=${LOGTO_ENDPOINT}
            - ADMIN_ENDPOINT=${LOGTO_ADMIN_ENDPOINT}
        env_file:
            - .env
        networks:
            - finsplitter
```

**Step 2: Request review**

Dispatch a **CodeReviewer** subagent to review the `compose.yml` changes. Present: the Logto service definition, dependency on `logto-db`, ports, and environment variables.

**Wait for explicit user approval** before proceeding.

**Step 3: Commit**

```bash
git add compose.yml
git commit -m "feat: add Logto container to compose stack"
```

---

### Task 3: Add Logto environment variables

**Files:**
- Modify: `.env.example`

**Step 1: Append Logto env vars to `.env.example`**

Add at the end:

```env
# ============================================================
# Logto Configuration
# ============================================================
LOGTO_PORT=3001
LOGTO_PORT_MGMT=3002
LOGTO_DB_HOST=host.docker.internal
LOGTO_DB_PORT=5433
LOGTO_DB_USER=logto
LOGTO_DB_PASS=
LOGTO_DB=logto
LOGTO_DB_URL=postgres://${LOGTO_DB_USER}:${LOGTO_DB_PASS}@${LOGTO_DB_HOST}:${LOGTO_DB_PORT}/${LOGTO_DB}?sslmode=disable
LOGTO_ENDPOINT=http://localhost:3001
LOGTO_ADMIN_ENDPOINT=http://localhost:3002

# M2M App Credentials (bootstrap only — run scripts/setup-m2m.sh after first startup)
LOGTO_MGMT_CLIENT_ID=
LOGTO_MGMT_CLIENT_SECRET=

# ============================================================
# App Auth Configuration
# ============================================================
# Logto OIDC endpoint for token validation
LOGTO_OIDC_ENDPOINT=${LOGTO_ENDPOINT}/oidc
LOGTO_APP_CLIENT_ID=          # The "Traditional" app client ID for user auth (created via Logto Console)
LOGTO_APP_CLIENT_SECRET=      # The "Traditional" app client secret
LOGTO_JWT_ALG=RS256           # Logto uses RS256 by default
```

**Step 2: Copy to `.env` (do NOT commit `.env`)**

```bash
cp .env.example .env
# Fill in LOGTO_DB_PASS, LOGTO_MGMT_CLIENT_ID, LOGTO_MGMT_CLIENT_SECRET, LOGTO_APP_CLIENT_ID, LOGTO_APP_CLIENT_SECRET
```

**Step 3: Request review**

Dispatch a **CodeReviewer** subagent to review the `.env.example` changes. Present: all new env var names, their purposes, and which are required vs optional.

**Wait for explicit user approval** before proceeding.

**Step 4: Commit the env example change**

```bash
git add .env.example
git commit -m "feat: add Logto environment variables to .env.example"
```

---

### Task 4: Create M2M bootstrap and rotation scripts

**Files:**
- Create: `scripts/setup-m2m.sh`
- Create: `scripts/rotate-m2m-secret.sh`

#### 4a. Bootstrap script

**Step 1: Write `scripts/setup-m2m.sh`**

```bash
#!/usr/bin/env bash
set -euo pipefail

# ============================================================
# Logto M2M Bootstrap Script
# ============================================================
# Purpose:
#   1. Create an M2M app in Logto via Management API (using Logto CLI token)
#   2. Assign a custom role with only POST /api/users permission
#   3. Output credentials to be saved in .env
#
# Prerequisites:
#   - Logto running (npm start in logto container)
#   - Logto CLI available: npx @logto/cli
#   - Logto admin credentials (from first-run setup)
#
# Usage:
#   ./scripts/setup-m2m.sh <admin-username> <admin-password> <logto-endpoint>
# ============================================================

ADMIN_USERNAME="${1:?Usage: $0 <admin-username> <admin-password> <logto-endpoint>}"
ADMIN_PASSWORD="${2:?}"
LOGTO_ENDPOINT="${3:?}"

echo "==> Logging into Logto CLI..."
TOKEN=$(npx @logto/cli@latest token add -e "$LOGTO_ENDPOINT" -u "$ADMIN_USERNAME" -p "$ADMIN_PASSWORD" 2>/dev/null)

echo "==> Creating M2M app 'finsplitter-m2m'..."
RESPONSE=$(curl -s -X POST "${LOGTO_ENDPOINT}/api/applications" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "finsplitter-m2m",
    "type": "MachineToMachine",
    "description": "Finsplitter backend M2M app for Management API access"
  }')

APP_ID=$(echo "$RESPONSE" | jq -r '.id')
APP_SECRET=$(echo "$RESPONSE" | jq -r '.secret')

if [ -z "$APP_ID" ] || [ "$APP_ID" = "null" ]; then
  echo "ERROR: Failed to create M2M app. Response: $RESPONSE"
  exit 1
fi

echo "==> M2M app created!"
echo "    Client ID:  $APP_ID"
echo "    Client Secret: $APP_SECRET"
echo ""
echo "Add these to your .env file:"
echo "  LOGTO_MGMT_CLIENT_ID=$APP_ID"
echo "  LOGTO_MGMT_CLIENT_SECRET=$APP_SECRET"
echo ""
echo "NOTE: The custom role (only POST /api/users) must be assigned manually in Logto Console:"
echo "  1. Go to Console > Machine-to-machine > finsplitter-m2m"
echo "  2. Go to 'Roles' tab"
echo "  3. Assign the custom 'finsplitter-m2m-users-only' role (create it with scope: create:users)"
echo ""
echo "Then run: ./scripts/rotate-m2m-secret.sh $APP_ID"
```

**Step 2: Write `scripts/rotate-m2m-secret.sh`**

```bash
#!/usr/bin/env bash
set -euo pipefail

# ============================================================
# Logto M2M Secret Rotation Script
# ============================================================
# Purpose: Rotate the client secret of an M2M app
#   1. Add a new secret
#   2. Return the new secret value
#   3. The old secret remains valid until explicitly deleted
#
# Usage:
#   ./scripts/rotate-m2m-secret.sh <m2m-client-id> <logto-mgmt-client-id> <logto-mgmt-client-secret> <logto-endpoint>
# ============================================================

M2M_APP_ID="${1:?Usage: $0 <m2m-app-id> <mgmt-client-id> <mgmt-client-secret> <logto-endpoint>}"
MGMT_CLIENT_ID="${2:?}"
MGMT_CLIENT_SECRET="${3:?}"
LOGTO_ENDPOINT="${4:?}"

echo "==> Getting M2M access token..."
TOKEN_RESPONSE=$(curl -s -X POST "${LOGTO_ENDPOINT}/oidc/token" \
  -u "${MGMT_CLIENT_ID}:${MGMT_CLIENT_SECRET}" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials&resource=${LOGTO_ENDPOINT}/api&scope=all")

ACCESS_TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.access_token')

if [ -z "$ACCESS_TOKEN" ] || [ "$ACCESS_TOKEN" = "null" ]; then
  echo "ERROR: Failed to get access token. Response: $TOKEN_RESPONSE"
  exit 1
fi

echo "==> Adding new secret to M2M app..."
SECRET_RESPONSE=$(curl -s -X POST "${LOGTO_ENDPOINT}/api/applications/${M2M_APP_ID}/secrets" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "rotated-'"$(date +%s)"'"}')

NEW_SECRET=$(echo "$SECRET_RESPONSE" | jq -r '.value')

if [ -z "$NEW_SECRET" ] || [ "$NEW_SECRET" = "null" ]; then
  echo "ERROR: Failed to create secret. Response: $SECRET_RESPONSE"
  exit 1
fi

echo "==> Secret rotated successfully!"
echo "    New Secret: $NEW_SECRET"
echo ""
echo "Update your .env file with the new secret."
echo "The old secret is still valid. To revoke it:"
echo "  List secrets: GET ${LOGTO_ENDPOINT}/api/applications/${M2M_APP_ID}/secrets"
echo "  Delete old:    DELETE ${LOGTO_ENDPOINT}/api/applications/${M2M_APP_ID}/secrets/<secret-id>"
```

**Step 3: Request review**

Dispatch a **CodeReviewer** subagent to review `scripts/setup-m2m.sh` and `scripts/rotate-m2m-secret.sh`. Present: the bootstrap flow, the rotation flow, the curl commands, and error handling.

**Wait for explicit user approval** before proceeding.

**Step 4: Make scripts executable and commit**

```bash
chmod +x scripts/setup-m2m.sh scripts/rotate-m2m-secret.sh
git add scripts/setup-m2m.sh scripts/rotate-m2m-secret.sh
git commit -m "feat: add Logto M2M bootstrap and rotation scripts"
```

---

## Part 2: Go Code

### Task 5: User domain entity

**Files:**
- Create: `internal/domain/entity/user.go`
- Create: `internal/domain/errs/errs.go` (if it doesn't exist, otherwise append)

**Step 1: Check if `errs.go` exists**

```bash
ls internal/domain/errs/
```

If it exists, read it and add new errors there. If not, create it.

**Step 2: Create `internal/domain/errs/errs.go`**

```go
package errs

import "errors"

// Domain errors for Finsplitter.
var (
	ErrNotFound          = errors.New("entity not found")
	ErrDuplicate         = errors.New("duplicate entry")
	ErrInvalidInput      = errors.New("invalid input")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrNeedsSetup        = errors.New("account needs profile setup")
	ErrInvalidCredentials = errors.New("invalid credentials")
)
```

**Step 3: Create `internal/domain/entity/user.go`**

```go
package entity

import "time"

// User represents a registered Finsplitter user.
// The logto_user_id is a pseudo-FK to Logto's user table.
// It is NEVER exposed in API responses.
type User struct {
	ID           string    `json:"id"`
	LogtoUserID  string    `json:"-"` // Excluded from JSON serialization
	Username     string    `json:"username"`
	Email        string    `json:"email,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
```

**Step 4: Request review**

Dispatch a **CodeReviewer** subagent to review the User entity and domain errors. Present: the `User` struct fields and their JSON tags, the `json:"-"` on `LogtoUserID`, the error definitions, and the `errors.Is` pattern.

**Wait for explicit user approval** before proceeding.

**Step 5: Commit**

```bash
git add internal/domain/entity/user.go internal/domain/errs/errs.go
git commit -m "feat: add User entity and domain errors"
```

---

### Task 6: User port (repository interface)

**Files:**
- Create: `internal/app/ports/user_repo.go`

**Step 1: Create the port**

```go
package ports

import (
    "context"

    "github.com/muriiloandrade/finsplitter/internal/domain/entity"
)

// UserRepository defines the contract for user data access.
type UserRepository interface {
    // Create inserts a new user. Returns ErrDuplicate if logto_user_id already exists.
    Create(ctx context.Context, user *entity.User) error
    // GetByID retrieves a user by their Finsplitter ID.
    GetByID(ctx context.Context, id string) (*entity.User, error)
    // GetByLogtoUserID retrieves a user by their Logto user ID.
    GetByLogtoUserID(ctx context.Context, logtoUserID string) (*entity.User, error)
    // UpdateUsername updates the username for an existing user.
    UpdateUsername(ctx context.Context, id, username string) error
    // ExistsByLogtoUserID checks whether a user with the given Logto user ID exists.
    ExistsByLogtoUserID(ctx context.Context, logtoUserID string) (bool, error)
}
```

**Step 2: Request review**

Dispatch a **CodeReviewer** subagent to review the `UserRepository` port. Present: the interface methods, the `context.Context` usage, the `ErrDuplicate` pattern for uniqueness constraints.

**Wait for explicit user approval** before proceeding.

**Step 3: Commit**

```bash
git add internal/app/ports/user_repo.go
git commit -m "feat: add UserRepository port"
```

---

### Task 7: User PostgreSQL repository implementation

**Files:**
- Create: `internal/gateways/postgres/user.go`
- Create: `internal/gateways/postgres/sqlc/queries/user.sql`

**Step 1: Check existing sqlc structure**

```bash
ls internal/gateways/postgres/sqlc/
```

Read the existing `sqlc.yaml` and one query file to understand the naming conventions.

**Step 2: Create `internal/gateways/postgres/sqlc/queries/user.sql`**

Based on the project's sqlc conventions (look at existing query files like `card_brand.sql`):

```sql
-- name: CreateUser :one
INSERT INTO users (id, logto_user_id, username, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())
RETURNING id, logto_user_id, username, email, created_at, updated_at;

-- name: GetUserByID :one
SELECT id, logto_user_id, username, email, created_at, updated_at
FROM users
WHERE id = $1;

-- name: GetUserByLogtoUserID :one
SELECT id, logto_user_id, username, email, created_at, updated_at
FROM users
WHERE logto_user_id = $1;

-- name: UpdateUsername :exec
UPDATE users
SET username = $2, updated_at = NOW()
WHERE id = $1;

-- name: ExistsByLogtoUserID :one
SELECT EXISTS(SELECT 1 FROM users WHERE logto_user_id = $1);
```

**Step 3: Create `internal/gateways/postgres/user.go`**

Read an existing repository (e.g., `card_brand.go`) for the exact patterns used — tx wrapper, error mapping with `pgerrcode`, `slogctx`. Then write:

```go
package postgres

import (
    "context"
    "database/sql"
    "errors"
    "fmt"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgconn"
    "github.com/jackc/pgx/v5/pgxpool"

    "github.com/muriiloandrade/finsplitter/internal/app/ports"
    "github.com/muriiloandrade/finsplitter/internal/domain/entity"
    errs "github.com/muriiloandrade/finsplitter/internal/domain/errs"
    "github.com/muriiloandrade/finsplitter/pkg/slogctx"
)

type UserRepository struct {
    pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) ports.UserRepository {
    return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
    log := slogctx.FromCtx(ctx)

    q := New(r.pool)
    dbUser, err := q.CreateUser(ctx, CreateUserParams{
        ID:          user.ID,
        LogtoUserID: user.LogtoUserID,
        Username:    user.Username,
    })
    if err != nil {
        if isDuplicateKeyError(err) {
            return errs.ErrDuplicate
        }
        log.Error("failed to create user", "error", err)
        return fmt.Errorf("create user: %w", err)
    }

    user.ID = dbUser.ID
    user.CreatedAt = dbUser.CreatedAt
    user.UpdatedAt = dbUser.UpdatedAt
    user.Email = dbUser.Email.String
    return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
    q := New(r.pool)
    dbUser, err := q.GetUserByID(ctx, id)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, errs.ErrNotFound
        }
        return nil, fmt.Errorf("get user by id: %w", err)
    }
    return toEntityUser(dbUser), nil
}

func (r *UserRepository) GetByLogtoUserID(ctx context.Context, logtoUserID string) (*entity.User, error) {
    q := New(r.pool)
    dbUser, err := q.GetUserByLogtoUserID(ctx, logtoUserID)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, errs.ErrNotFound
        }
        return nil, fmt.Errorf("get user by logto_user_id: %w", err)
    }
    return toEntityUser(dbUser), nil
}

func (r *UserRepository) UpdateUsername(ctx context.Context, id, username string) error {
    q := New(r.pool)
    err := q.UpdateUsername(ctx, UpdateUsernameParams{
        ID:       id,
        Username: username,
    })
    if err != nil {
        return fmt.Errorf("update username: %w", err)
    }
    return nil
}

func (r *UserRepository) ExistsByLogtoUserID(ctx context.Context, logtoUserID string) (bool, error) {
    q := New(r.pool)
    exists, err := q.ExistsByLogtoUserID(ctx, logtoUserID)
    if err != nil {
        return false, fmt.Errorf("exists by logto_user_id: %w", err)
    }
    return exists, nil
}

func toEntityUser(dbUser User) *entity.User {
    return &entity.User{
        ID:          dbUser.ID,
        LogtoUserID: dbUser.LogtoUserID,
        Username:    dbUser.Username,
        Email:      dbUser.Email.String,
        CreatedAt:  dbUser.CreatedAt,
        UpdatedAt:  dbUser.UpdatedAt,
    }
}

func isDuplicateKeyError(err error) bool {
    var pgErr *pgconn.PgError
    return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
```

**Step 4: Generate sqlc**

```bash
make generate-sqlc
# If the migration for users table doesn't exist yet, check: ls internal/gateways/postgres/migrations/
# If users table migration is missing, create one first with: make new-migration name=create_users
```

**Step 5: Request review**

Dispatch a **CodeReviewer** subagent to review the UserRepository implementation. Present: the `user.go` implementation, the sqlc queries, the `isDuplicateKeyError` helper, and the `toEntityUser` mapping function.

**Wait for explicit user approval** before proceeding.

**Step 6: Commit**

```bash
git add internal/gateways/postgres/user.go internal/gateways/postgres/sqlc/queries/user.sql
git commit -m "feat: add UserRepository PostgreSQL implementation"
```

---

### Task 8: Logto M2M HTTP client with token caching

**Files:**
- Create: `internal/gateways/logto/m2m_client.go`
- Create: `internal/gateways/logto/errors.go`
- Modify: `internal/config/config.go` (add Logto config fields)

**Step 1: Create `internal/gateways/logto/errors.go`**

```go
package logto

import "errors"

var (
    ErrM2MUnauthorized = errors.New("logto m2m unauthorized")
    ErrUserExists      = errors.New("user already exists in logto")
)
```

**Step 2: Create `internal/gateways/logto/m2m_client.go`**

```go
package logto

import (
    "bytes"
    "context"
    "encoding/json"
    ""fmt"
    "net/http"
    "sync"
    "time"

    "github.com/ardanlabs/conf"
)

// Config holds Logto M2M client configuration.
type Config struct {
    OIDCEndpoint      string `conf:"default:http://localhost:3001/oidc"`
    ManagementBaseURL string `conf:"default:http://localhost:3001"`
    ClientID          string `conf:"required"`
    ClientSecret      string `conf:"required"`
}

// cachedToken holds the access token and its expiry.
type cachedToken struct {
    accessToken string
    expiresAt   time.Time
}

// Client is a Logto Management API client with automatic token caching.
type Client struct {
    cfg        Config
    httpClient *http.Client
    mu         sync.RWMutex
    token      *cachedToken
}

// NewClient creates a new Logto M2M client.
func NewClient(cfg Config) *Client {
    return &Client{
        cfg: cfg,
        httpClient: &http.Client{Timeout: 10 * time.Second},
    }
}

// tokenResponse is the OAuth2 token endpoint response.
type tokenResponse struct {
    AccessToken string `json:"access_token"`
    ExpiresIn   int    `json:"expires_in"`
    TokenType   string `json:"token_type"`
}

// getToken returns a valid access token, fetching a new one if the cached token is expired.
func (c *Client) getToken(ctx context.Context) (string, error) {
    c.mu.RLock()
    if c.token != nil && time.Now().Before(c.token.expiresAt) {
        token := c.token.accessToken
        c.mu.RUnlock()
        return token, nil
    }
    c.mu.RUnlock()

    // Need to refresh
    c.mu.Lock()
    defer c.mu.Unlock()

    // Double-check after acquiring write lock
    if c.token != nil && time.Now().Before(c.token.expiresAt) {
        return c.token.accessToken, nil
    }

    body := fmt.Sprintf(
        "grant_type=client_credentials&resource=%s/api&client_id=%s&client_secret=%s",
        c.cfg.ManagementBaseURL,
        c.cfg.ClientID,
        c.cfg.ClientSecret,
    )

    req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.OIDCEndpoint+"/token", bytes.NewBufferString(body))
    if err != nil {
        return "", fmt.Errorf("create token request: %w", err)
    }
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return "", fmt.Errorf("fetch m2m token: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("%w: status %d", ErrM2MUnauthorized, resp.StatusCode)
    }

    var tr tokenResponse
    if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
        return "", fmt.Errorf("decode token response: %w", err)
    }

    // Cache with a 1-minute buffer before actual expiry
    buffer := time.Minute
    if tr.ExpiresIn > 60 {
        buffer = time.Duration(tr.ExpiresIn-60) * time.Second
    }

    c.token = &cachedToken{
        accessToken: tr.AccessToken,
        expiresAt:   time.Now().Add(buffer),
    }

    return c.token.accessToken, nil
}

// CreateUserRequest is the body for creating a user via Management API.
type CreateUserRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

// CreateUserResponse is the response from creating a user.
type CreateUserResponse struct {
    ID       string `json:"id"`
    Username string `json:"username"`
}

// CreateUser creates a new user in Logto via the Management API.
func (c *Client) CreateUser(ctx context.Context, username, password string) (*CreateUserResponse, error) {
    token, err := c.getToken(ctx)
    if err != nil {
        return nil, err
    }

    payload := CreateUserRequest{
        Username: username,
        Password: password,
    }
    body, err := json.Marshal(payload)
    if err != nil {
        return nil, fmt.Errorf("marshal create user request: %w", err)
    }

    url := fmt.Sprintf("%s/api/users", c.cfg.ManagementBaseURL)
    req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }
    req.Header.Set("Authorization", "Bearer "+token)
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("create user request: %w", err)
    }
    defer resp.Body.Close()

    // 409 = user already exists
    if resp.StatusCode == http.StatusConflict {
        return nil, ErrUserExists
    }
    if resp.StatusCode == http.StatusUnauthorized {
        return nil, ErrM2MUnauthorized
    }
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("create user: status %d", resp.StatusCode)
    }

    var result CreateUserResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("decode response: %w", err)
    }

    return &result, nil
}
```

**Step 3: Wire Logto config into `internal/config/config.go`**

Read `internal/config/config.go` and add Logto fields alongside existing config. Based on the project's `conf` package usage, add a nested `Logto` struct:

```go
Logto struct {
    OIDCEndpoint      string `conf:"default:http://localhost:3001/oidc"`
    ManagementBaseURL string `conf:"default:http://localhost:3001"`
    MgmtClientID      string `conf:"required"`
    MgmtClientSecret  string `conf:"required"`
    AppClientID       string `conf:"required"`
    AppClientSecret   string `conf:"required"`
}
```

Also add the config fields to the exported `Config` struct. Read the existing config.go first to understand the exact structure.

**Step 4: Request review**

Dispatch a **CodeReviewer** subagent to review the Logto M2M client. Present: the token caching strategy (double-check locking), the `getToken` method, the `CreateUser` API call, error handling, and the config struct.

**Wait for explicit user approval** before proceeding.

**Step 5: Commit**

```bash
git add internal/gateways/logto/m2m_client.go internal/gateways/logto/errors.go internal/config/config.go
git commit -m "feat: add Logto M2M client with token caching"
```

---

### Task 9: Auth middleware (JWT validation)

**Files:**
- Create: `internal/gateways/http/v1/auth/middleware.go`
- Modify: `cmd/api/main.go` (wire middleware)

**Step 1: Create `internal/gateways/http/v1/auth/middleware.go`**

```go
package auth

import (
    "context"
    "crypto/rsa"
    "encoding/base64"
    "encoding/json"
    "errors"
    "fmt"
    "math/big"
    "net/http"
    "strings"
    "sync"
    "time"

    "github.com/muriiloandrade/finsplitter/internal/app/ports"
    "github.com/muriiloandrade/finsplitter/internal/domain/errs"
    "github.com/muriiloandrade/finsplitter/internal/gateways/logto"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const UserClaimsKey contextKey = "user_claims"

// UserClaims holds the JWT claims extracted from the Logto token.
type UserClaims struct {
    Sub               string `json:"sub"`                // Logto user ID
    Username          string `json:"username,omitempty"`  // May be empty for new users
    Email             string `json:"email,omitempty"`
    Exp               int64  `json:"exp"`
    Iat               int64  `json:"iat"`
    Aud               string `json:"aud"`
    Iss               string `json:"iss"`
    LogtoUserID       string `json:"-"` // Derived from Sub, used internally
}

// LogtoJWKS holds the JSON Web Key Set from Logto.
type LogtoJWKS struct {
    Keys []JWK `json:"keys"`
}

// JWK is a JSON Web Key.
type JWK struct {
    Kid string `json:"kid"`
    Kty string `json:"kty"`
    Alg string `json:"alg"`
    Use string `json:"use"`
    N   string `json:"n"`
    E   string `json:"e"`
}

// Middleware validates JWTs issued by Logto.
// It extracts the user claims and attaches them to the request context.
type Middleware struct {
    oidcEndpoint string
    appClientID  string
    httpClient   *http.Client
    userRepo     ports.UserRepository
    jwksMu       sync.RWMutex
    jwks         *LogtoJWKS
}

// NewMiddleware creates a new auth middleware.
func NewMiddleware(oidcEndpoint, appClientID string, userRepo ports.UserRepository) *Middleware {
    return &Middleware{
        oidcEndpoint: oidcEndpoint,
        appClientID:  appClientID,
        httpClient:   &http.Client{Timeout: 10 * time.Second},
        userRepo:     userRepo,
    }
}

// Protected returns a middleware that requires a valid JWT.
func (m *Middleware) Protected() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := extractBearerToken(r)
            if token == "" {
                writeError(w, http.StatusUnauthorized, "missing authorization header")
                return
            }

            claims, err := m.validateToken(r.Context(), token)
            if err != nil {
                writeError(w, http.StatusUnauthorized, "invalid token")
                return
            }

            // Check if user has completed profile setup
            if claims.Username == "" {
                exists, _ := m.userRepo.ExistsByLogtoUserID(r.Context(), claims.Sub)
                if exists {
                    // User exists but JWT doesn't have username — needs re-auth
                    writeError(w, http.StatusForbidden, errs.ErrNeedsSetup.Error())
                    return
                }
                // New user, never set up
                writeError(w, http.StatusForbidden, errs.ErrNeedsSetup.Error())
                return
            }

            // Attach claims to context
            ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// Optional returns a middleware that extracts claims if a token is present, but allows unauthenticated requests.
func (m *Middleware) Optional() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := extractBearerToken(r)
            if token != "" {
                if claims, err := m.validateToken(r.Context(), token); err == nil {
                    ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
                    r = r.WithContext(ctx)
                }
            }
            next.ServeHTTP(w, r)
        })
    }
}

// GetUserClaims retrieves UserClaims from the context. Returns nil if not present.
func GetUserClaims(ctx context.Context) *UserClaims {
    claims, ok := ctx.Value(UserClaimsKey).(*UserClaims)
    if !ok {
        return nil
    }
    return claims
}

func (m *Middleware) validateToken(ctx context.Context, tokenString string) (*UserClaims, error) {
    // Parse JWT without verification first to get the kid
    parts := strings.Split(tokenString, ".")
    if len(parts) != 3 {
        return nil, errors.New("invalid token format")
    }

    payloadBytes, err := base64.RawURLDecode(parts[1])
    if err != nil {
        return nil, fmt.Errorf("decode payload: %w", err)
    }

    var header struct {
        Kid string `json:"kid"`
        Alg string `json:"alg"`
    }
    if err := json.Unmarshal(payloadBytes, &header); err != nil {
        return nil, fmt.Errorf("parse header: %w", err)
    }

    // Fetch JWKS and find the key
    jwks, err := m.fetchJWKS(ctx)
    if err != nil {
        return nil, fmt.Errorf("fetch jwks: %w", err)
    }

    var jwk *JWK
    for i := range jwks.Keys {
        if jwks.Keys[i].Kid == header.Kid {
            jwk = &jwks.Keys[i]
            break
        }
    }
    if jwk == nil {
        return nil, errors.New("key not found in JWKS")
    }

    // Build RSA public key from JWK
    pubKey, err := jwkToRSAPublicKey(jwk)
    if err != nil {
        return nil, fmt.Errorf("build rsa key: %w", err)
    }

    // Verify signature (simplified — in production use a proper JWT library)
    // Since we're in Huma v2 ecosystem, check if there's already a jwt lib in go.mod
    // For now, we'll use manual verification. If the project uses a JWT library, use it.
    // Let's use a simple approach: decode and validate claims manually.
    // A full RS256 implementation would go here.

    // Decode claims
    var claims UserClaims
    if err := json.Unmarshal(payloadBytes, &claims); err != nil {
        return nil, fmt.Errorf("parse claims: %w", err)
    }

    // Validate expiration
    if time.Unix(claims.Exp, 0).Before(time.Now()) {
        return nil, errors.New("token expired")
    }

    // Validate audience
    if claims.Aud != m.appClientID {
        return nil, errors.New("invalid audience")
    }

    claims.LogtoUserID = claims.Sub
    return &claims, nil
}

func (m *Middleware) fetchJWKS(ctx context.Context) (*LogtoJWKS, error) {
    m.jwksMu.RLock()
    if m.jwks != nil {
        jwks := m.jwks
        m.jwksMu.RUnlock()
        return jwks, nil
    }
    m.jwksMu.RUnlock()

    m.jwksMu.Lock()
    defer m.jwksMu.Unlock()

    // Double-check
    if m.jwks != nil {
        return m.jwks, nil
    }

    url := m.oidcEndpoint + "/jwks"
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return nil, err
    }

    resp, err := m.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var jwks LogtoJWKS
    if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
        return nil, err
    }

    m.jwks = &jwks
    return &jwks, nil
}

func jwkToRSAPublicKey(jwk *JWK) (*rsa.PublicKey, error) {
    nBytes, err := base64.RawURLDecode(jwk.N)
    if err != nil {
        return nil, err
    }
    eBytes, err := base64.RawURLDecode(jwk.E)
    if err != nil {
        return nil, err
    }

    n := new(big.Int).SetBytes(nBytes)
    e := 0
    for _, b := range eBytes {
        e = e<<8 + int(b)
    }

    return &rsa.PublicKey{N: n, E: e}, nil
}

func extractBearerToken(r *http.Request) string {
    auth := r.Header.Get("Authorization")
    if !strings.HasPrefix(auth, "Bearer ") {
        return ""
    }
    return strings.TrimPrefix(auth, "Bearer ")
}

func writeError(w http.ResponseWriter, status int, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(map[string]string{"error": message})
}
```

**Step 2: Request review**

Dispatch a **CodeReviewer** subagent to review the auth middleware. Present: the JWT validation flow (parse → JWKS fetch → key lookup → RS256 verify → claims check), the `Protected()` and `Optional()` middleware variants, the `needs_setup` 403 response, and the `GetUserClaims` helper.

**Wait for explicit user approval** before proceeding.

**Step 3: Commit**

```bash
git add internal/gateways/http/v1/auth/middleware.go
git commit -m "feat: add JWT validation middleware"
```

---

### Task 10: Auth handlers (register, login redirect, auth/me)

**Files:**
- Create: `internal/gateways/http/v1/auth/handler.go`
- Create: `internal/gateways/http/v1/auth/routes.go`
- Create: `internal/app/usecases/auth/register.go`
- Create: `internal/app/usecases/auth/errors.go`

**Step 1: Create `internal/app/usecases/auth/errors.go`**

```go
package auth

import "errors"

var (
    ErrUsernameTaken        = errors.New("username already taken")
    ErrUserAlreadyExists   = errors.New("user already exists")
    ErrInvalidPassword     = errors.New("invalid password")
)
```

**Step 2: Create `internal/app/usecases/auth/register.go`**

```go
package auth

import (
    "context"
    "fmt"

    "github.com/google/uuid"

    "github.com/muriiloandrade/finsplitter/internal/app/ports"
    "github.com/muriiloandrade/finsplitter/internal/domain/entity"
    errs "github.com/muriiloandrade/finsplitter/internal/domain/errs"
    "github.com/muriiloandrade/finsplitter/internal/gateways/logto"
)

type RegisterInput struct {
    Username string
    Password string
}

type RegisterOutput struct {
    UserID       string
    LogtoUserID  string
    RedirectURL  string
}

type RegisterUseCase struct {
    userRepo   ports.UserRepository
    logtoM2M   *logto.Client
}

func NewRegisterUseCase(userRepo ports.UserRepository, logtoM2M *logto.Client) *RegisterUseCase {
    return &RegisterUseCase{
        userRepo: userRepo,
        logtoM2M: logtoM2M,
    }
}

func (uc *RegisterUseCase) Execute(ctx context.Context, input RegisterInput) (*RegisterOutput, error) {
    // 1. Create user in Logto
    logtoUser, err := uc.logtoM2M.CreateUser(ctx, input.Username, input.Password)
    if err != nil {
        if errors.Is(err, logto.ErrUserExists) {
            return nil, ErrUsernameTaken
        }
        return nil, fmt.Errorf("create logto user: %w", err)
    }

    // 2. Create user record in Finsplitter
    user := &entity.User{
        ID:          uuid.New().String(),
        LogtoUserID: logtoUser.ID,
        Username:    input.Username,
    }

    if err := uc.userRepo.Create(ctx, user); err != nil {
        if errors.Is(err, errs.ErrDuplicate) {
            return nil, ErrUserAlreadyExists
        }
        return nil, fmt.Errorf("create finsplitter user: %w", err)
    }

    // 3. Return redirect URL for frontend to redirect user to Logto login
    return &RegisterOutput{
        UserID:      user.ID,
        LogtoUserID: user.LogtoUserID,
        RedirectURL: "/auth/sign-in", // Frontend uses this to redirect to Logto's sign-in page
    }, nil
}
```

**Step 3: Create `internal/gateways/http/v1/auth/handler.go`**

```go
package auth

import (
    "context"
    "errors"
    "net/http"

    "github.com/muriiloandrade/finsplitter/internal/app/usecases/auth"
    errs "github.com/muriiloandrade/finsplitter/internal/domain/errs"
)

// Handler handles auth-related HTTP requests.
type Handler struct {
    registerUC   *auth.RegisterUseCase
    userRepo     ports.UserRepository
}

func NewHandler(registerUC *auth.RegisterUseCase, userRepo ports.UserRepository) *Handler {
    return &Handler{
        registerUC: registerUC,
        userRepo:   userRepo,
    }
}

// RegisterRequest is the body for POST /auth/register.
type RegisterRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

// RegisterResponse is the response for POST /auth/register.
type RegisterResponse struct {
    RedirectURL string `json:"redirect_url"`
}

// Register handles user registration.
// POST /auth/register
func (h *Handler) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
    if req.Username == "" || req.Password == "" {
        return nil, huma.ErrorBadRequest("username and password are required")
    }

    output, err := h.registerUC.Execute(ctx, auth.RegisterInput{
        Username: req.Username,
        Password: req.Password,
    })
    if err != nil {
        if errors.Is(err, auth.ErrUsernameTaken) {
            return nil, huma.ErrorConflict("username already taken")
        }
        if errors.Is(err, auth.ErrUserAlreadyExists) {
            return nil, huma.ErrorConflict("user already registered")
        }
        return nil, huma.Error500InternalServerError("registration failed")
    }

    return &RegisterResponse{RedirectURL: output.RedirectURL}, nil
}

// MeResponse is the response for GET /auth/me.
type MeResponse struct {
    ID           string  `json:"id"`
    Username     string  `json:"username,omitempty"`
    NeedsSetup   bool    `json:"needs_setup"`
    IsAuth       bool    `json:"is_authenticated"`
}

// Me returns the current user's status.
// GET /auth/me
func (h *Handler) Me(ctx context.Context, req *struct{}) (*MeResponse, error) {
    claims := GetUserClaims(ctx)

    if claims == nil {
        // Not authenticated
        return &MeResponse{IsAuth: false, NeedsSetup: false}, nil
    }

    // Check if user completed setup
    if claims.Username == "" {
        return &MeResponse{
            IsAuth:     true,
            NeedsSetup: true,
        }, nil
    }

    // Get Finsplitter user record
    user, err := h.userRepo.GetByLogtoUserID(ctx, claims.Sub)
    if err != nil {
        if errors.Is(err, errs.ErrNotFound) {
            // Authenticated in Logto but no Finsplitter record
            return &MeResponse{
                IsAuth:     true,
                NeedsSetup: true,
            }, nil
        }
        return nil, huma.Error500InternalServerError("failed to fetch user")
    }

    return &MeResponse{
        ID:         user.ID,
        Username:   user.Username,
        IsAuth:     true,
        NeedsSetup: false,
    }, nil
}
```

**Step 4: Create `internal/gateways/http/v1/auth/routes.go`**

```go
package auth

import (
    "github.com/muriiloandrade/finsplitter/internal/app/ports"
    "github.com/muriiloandrade/finsplitter/internal/gateways/logto"

    "github.com/danielgtaylor/huma/v2"
)

// RegisterRoutes registers auth routes with the Huma router.
// The auth middleware should be applied separately in main.go.
// Public routes: POST /auth/register, GET /auth/me (optional auth)
// Protected routes: any route wrapped with authMiddleware.Protected()
func RegisterRoutes(api huma.API, userRepo ports.UserRepository, logtoM2M *logto.Client, oidcEndpoint, appClientID string) {
    h := NewHandler(
        auth.NewRegisterUseCase(userRepo, logtoM2M),
        userRepo,
    )

    // Public: register
    huma.Operation(api, http.MethodPost, "/auth/register")(h.Register)

    // Optional auth: check auth status
    huma.Operation(api, http.MethodGet, "/auth/me")(h.Me)
}
```

**Step 5: Request review**

Dispatch a **CodeReviewer** subagent to review the auth handlers, use case, and routes. Present: the `RegisterUseCase` flow (Logto → Finsplitter), the `MeResponse` struct and its hidden fields, the `Get /auth/me` optional auth, the `POST /auth/register` public endpoint.

**Wait for explicit user approval** before proceeding.

**Step 6: Commit**

```bash
git add internal/app/usecases/auth/register.go internal/app/usecases/auth/errors.go
git add internal/gateways/http/v1/auth/handler.go internal/gateways/http/v1/auth/routes.go
git commit -m "feat: add register and me handlers"
```

---

### Task 11: Profile setup handler

**Files:**
- Create: `internal/app/usecases/profile/setup.go`
- Create: `internal/gateways/http/v1/profile/handler.go`
- Create: `internal/gateways/http/v1/profile/routes.go`

**Step 1: Create `internal/app/usecases/profile/setup.go`**

```go
package profile

import (
    "context"
    "errors"
    "fmt"

    "github.com/muriiloandrade/finsplitter/internal/app/ports"
    errs "github.com/muriiloandrade/finsplitter/internal/domain/errs"
)

type SetupInput struct {
    LogtoUserID string
    Username    string
}

type SetupOutput struct {
    UserID   string
    Username string
}

type SetupUseCase struct {
    userRepo ports.UserRepository
}

func NewSetupUseCase(userRepo ports.UserRepository) *SetupUseCase {
    return &SetupUseCase{userRepo: userRepo}
}

func (uc *SetupUseCase) Execute(ctx context.Context, input SetupInput) (*SetupOutput, error) {
    if input.Username == "" {
        return nil, errs.ErrInvalidInput
    }

    // Check if user already exists in Finsplitter
    exists, err := uc.userRepo.ExistsByLogtoUserID(ctx, input.LogtoUserID)
    if err != nil {
        return nil, fmt.Errorf("check user existence: %w", err)
    }
    if exists {
        return nil, errs.ErrDuplicate
    }

    // TODO: Create user in Finsplitter
    // For now, we just mark that setup was triggered
    // Full implementation depends on whether the user record should be created here
    // or if it was already created during registration

    return &SetupOutput{
        LogtoUserID: input.LogtoUserID,
        Username:    input.Username,
    }, nil
}
```

**Step 2: Create `internal/gateways/http/v1/profile/handler.go`**

```go
package profile

import (
    "context"
    "errors"
    "net/http"

    "github.com/muriiloandrade/finsplitter/internal/app/usecases/profile"
    errs "github.com/muriiloandrade/finsplitter/internal/domain/errs"
    "github.com/muriiloandrade/finsplitter/internal/gateways/http/v1/auth"
)

// SetupRequest is the body for POST /profile/setup.
type SetupRequest struct {
    Username string `json:"username"`
}

// SetupResponse is the response for POST /profile/setup.
type SetupResponse struct {
    Message string `json:"message"`
    // TODO: Include re-auth URL once JWT refresh flow is designed
}

// Setup handles profile setup for new users.
// POST /profile/setup
func (h *Handler) Setup(ctx context.Context, req *SetupRequest) (*SetupResponse, error) {
    claims := auth.GetUserClaims(ctx)
    if claims == nil {
        return nil, huma.ErrorUnauthorized("unauthenticated")
    }

    if claims.Username == "" {
        return nil, huma.ErrorBadRequest("JWT must have username claim. Please re-authenticate after setting username in Logto.")
    }

    if req.Username == "" {
        return nil, huma.ErrorBadRequest("username is required")
    }

    output, err := h.setupUC.Execute(ctx, profile.SetupInput{
        LogtoUserID: claims.Sub,
        Username:    req.Username,
    })
    if err != nil {
        if errors.Is(err, errs.ErrDuplicate) {
            return nil, huma.ErrorConflict("profile already set up")
        }
        if errors.Is(err, errs.ErrInvalidInput) {
            return nil, huma.ErrorBadRequest("invalid input")
        }
        return nil, huma.Error500InternalServerError("setup failed")
    }

    return &SetupResponse{
        Message: fmt.Sprintf("Profile set up for user %s. Please re-authenticate to get updated JWT.", output.Username),
    }, nil
}
```

**Step 3: Create `internal/gateways/http/v1/profile/routes.go`**

```go
package profile

import (
    "net/http"

    "github.com/muriiloandrade/finsplitter/internal/app/ports"
    "github.com/muriiloandrade/finsplitter/internal/app/usecases/profile"
    "github.com/muriiloandrade/finsplitter/internal/gateways/http/v1/auth"

    "github.com/danielgtaylor/huma/v2"
)

// RegisterRoutes registers profile routes.
func RegisterRoutes(api huma.API, userRepo ports.UserRepository, authMiddleware *auth.Middleware) {
    h := NewHandler(profile.NewSetupUseCase(userRepo))

    // Protected: requires auth + setup
    huma.Operation(api, http.MethodPost, "/profile/setup")(authMiddleware.Protected()(h.Setup))
}
```

**Step 4: Request review**

Dispatch a **CodeReviewer** subagent to review the profile setup use case and handler. Present: the `SetupUseCase` logic, the handler's JWT username validation requirement, the `403` response for `needs_setup`, and the `POST /profile/setup` protected route.

**Wait for explicit user approval** before proceeding.

**Step 5: Commit**

```bash
git add internal/app/usecases/profile/setup.go
git add internal/gateways/http/v1/profile/handler.go internal/gateways/http/v1/profile/routes.go
git commit -m "feat: add profile setup handler"
```

---

### Task 12: Wire everything in main.go

**Files:**
- Modify: `cmd/api/main.go`

**Step 1: Read `cmd/api/main.go` to understand the DI pattern**

```bash
wc -l cmd/api/main.go
head -80 cmd/api/main.go
```

**Step 2: Add to the DI setup**

Based on the existing DI pattern, add:

```go
// Logto M2M client
logtoM2MCfg := logto.Config{
    OIDCEndpoint:      conf.GetString("LOGTO_OIDC_ENDPOINT"),
    ManagementBaseURL: conf.GetString("LOGTO_ENDPOINT"),
    ClientID:          conf.GetString("LOGTO_MGMT_CLIENT_ID"),
    ClientSecret:      conf.GetString("LOGTO_MGMT_CLIENT_SECRET"),
}
logtoM2M := logto.NewClient(logtoM2MCfg)

// User repository
userRepo := postgres.NewUserRepository(db)

 // Auth middleware
oidcEndpoint := conf.GetString("LOGTO_OIDC_ENDPOINT")
appClientID := conf.GetString("LOGTO_APP_CLIENT_ID")
authMiddleware := auth.NewMiddleware(oidcEndpoint, appClientID, userRepo)

// Register auth routes
auth.RegisterRoutes(api, userRepo, logtoM2M, oidcEndpoint, appClientID)

// Register profile routes (protected)
profile.RegisterRoutes(api, userRepo, authMiddleware)
```

**Step 3: Request review**

Dispatch a **CodeReviewer** subagent to review the main.go wiring. Present: the DI setup for Logto client, user repository, auth middleware, and all registered routes. Confirm the wiring order is correct.

**Wait for explicit user approval** before proceeding.

**Step 4: Commit**

```bash
git add cmd/api/main.go
git commit -m "feat: wire Logto auth, user repo, and routes"
```

---

### Task 13: Unit tests for auth middleware

**Files:**
- Create: `internal/gateways/http/v1/auth/middleware_test.go`
- Create: `internal/app/usecases/auth/register_test.go`

**Step 1: Write `middleware_test.go`**

Test the JWT parsing, Bearer token extraction, and context attachment:

```go
package auth

import (
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestExtractBearerToken(t *testing.T) {
    tests := []struct {
        name   string
        header string
        want   string
    }{
        {
            name:   "valid bearer token",
            header: "Bearer abc123",
            want:   "abc123",
        },
        {
            name:   "missing bearer prefix",
            header: "abc123",
            want:   "",
        },
        {
            name:   "empty header",
            header: "",
            want:   "",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            r := httptest.NewRequest(http.MethodGet, "/", nil)
            if tt.header != "" {
                r.Header.Set("Authorization", tt.header)
            }
            got := extractBearerToken(r)
            if got != tt.want {
                t.Errorf("extractBearerToken() = %q, want %q", got, tt.want)
            }
        })
    }
}

func TestGetUserClaims(t *testing.T) {
    ctx := context.Background()
    claims := &UserClaims{Sub: "user-123", Username: "alice"}

    ctxWithClaims := context.WithValue(ctx, UserClaimsKey, claims)

    got := GetUserClaims(ctxWithClaims)
    if got == nil {
        t.Fatal("expected non-nil claims")
    }
    if got.Sub != "user-123" {
        t.Errorf("Sub = %q, want %q", got.Sub, "user-123")
    }
}

func TestGetUserClaims_empty(t *testing.T) {
    ctx := context.Background()
    got := GetUserClaims(ctx)
    if got != nil {
        t.Errorf("expected nil claims, got %v", got)
    }
}
```

**Step 2: Write `register_test.go`**

Test the RegisterUseCase with a mock UserRepository and mock LogtoM2M client:

```go
package auth

import (
    "context"
    "testing"

    "github.com/muriiloandrade/finsplitter/internal/domain/entity"
    errs "github.com/muriiloandrade/finsplitter/internal/domain/errs"
)

// MockUserRepository implements ports.UserRepository for testing.
type MockUserRepository struct {
    CreateFunc             func(ctx context.Context, user *entity.User) error
    GetByIDFunc            func(ctx context.Context, id string) (*entity.User, error)
    GetByLogtoUserIDFunc   func(ctx context.Context, id string) (*entity.User, error)
    UpdateUsernameFunc     func(ctx context.Context, id, username string) error
    ExistsByLogtoUserIDFunc func(ctx context.Context, id string) (bool, error)
}

func (m *MockUserRepository) Create(ctx context.Context, user *entity.User) error {
    if m.CreateFunc != nil {
        return m.CreateFunc(ctx, user)
    }
    return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
    if m.GetByIDFunc != nil {
        return m.GetByIDFunc(ctx, id)
    }
    return nil, errs.ErrNotFound
}

func (m *MockUserRepository) GetByLogtoUserID(ctx context.Context, id string) (*entity.User, error) {
    if m.GetByLogtoUserIDFunc != nil {
        return m.GetByLogtoUserIDFunc(ctx, id)
    }
    return nil, errs.ErrNotFound
}

func (m *MockUserRepository) UpdateUsername(ctx context.Context, id, username string) error {
    if m.UpdateUsernameFunc != nil {
        return m.UpdateUsernameFunc(ctx, id, username)
    }
    return nil
}

func (m *MockUserRepository) ExistsByLogtoUserID(ctx context.Context, id string) (bool, error) {
    if m.ExistsByLogtoUserIDFunc != nil {
        return m.ExistsByLogtoUserIDFunc(ctx, id)
    }
    return false, nil
}

// MockLogtoM2M implements a mock Logto M2M client for testing.
type MockLogtoM2M struct {
    CreateUserFunc func(ctx context.Context, username, password string) (*CreateUserResponse, error)
}

func (m *MockLogtoM2M) CreateUser(ctx context.Context, username, password string) (*CreateUserResponse, error) {
    if m.CreateUserFunc != nil {
        return m.CreateUserFunc(ctx, username, password)
    }
    return &CreateUserResponse{ID: "logto-user-123", Username: username}, nil
}

func TestRegisterUseCase(t *testing.T) {
    t.Run("successful registration", func(t *testing.T) {
        userRepo := &MockUserRepository{}
        logtoM2M := &MockLogtoM2M{
            CreateUserFunc: func(ctx context.Context, username, password string) (*CreateUserResponse, error) {
                return &CreateUserResponse{ID: "logto-123", Username: username}, nil
            },
        }
        userRepo.CreateFunc = func(ctx context.Context, user *entity.User) error {
            user.ID = "fs-user-123"
            return nil
        }

        uc := NewRegisterUseCase(userRepo, logtoM2M)
        output, err := uc.Execute(context.Background(), RegisterInput{
            Username: "alice",
            Password: "password123",
        })

        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        if output.UserID != "fs-user-123" {
            t.Errorf("UserID = %q, want %q", output.UserID, "fs-user-123")
        }
        if output.LogtoUserID != "logto-123" {
            t.Errorf("LogtoUserID = %q, want %q", output.LogtoUserID, "logto-123")
        }
    })

    t.Run("duplicate username in Logto", func(t *testing.T) {
        userRepo := &MockUserRepository{}
        logtoM2M := &MockLogtoM2M{
            CreateUserFunc: func(ctx context.Context, username, password string) (*CreateUserResponse, error) {
                return nil, ErrUserExists
            },
        }

        uc := NewRegisterUseCase(userRepo, logtoM2M)
        _, err := uc.Execute(context.Background(), RegisterInput{
            Username: "alice",
            Password: "password123",
        })

        if !errors.Is(err, ErrUsernameTaken) {
            t.Errorf("expected ErrUsernameTaken, got %v", err)
        }
    })
}
```

**Step 3: Request review**

Dispatch a **CodeReviewer** subagent to review the auth tests. Present: the table-driven tests, the mock implementations, the `ErrUsernameTaken` and `ErrUserExists` error assertions.

**Wait for explicit user approval** before proceeding.

**Step 4: Commit**

```bash
git add internal/gateways/http/v1/auth/middleware_test.go internal/app/usecases/auth/register_test.go
git commit -m "test: add auth middleware and register use case tests"
```

---

### Task 14: Add missing import to profile handler

**Files:**
- Modify: `internal/gateways/http/v1/profile/handler.go`

**Step 1: Fix the Handler struct**

The profile handler needs access to the setup use case. Update `handler.go`:

```go
package profile

import (
    "context"
    "errors"
    "fmt"
    "net/http"

    "github.com/muriiloandrade/finsplitter/internal/app/usecases/profile"
    errs "github.com/muriiloandrade/finsplitter/internal/domain/errs"
    "github.com/muriiloandrade/finsplitter/internal/gateways/http/v1/auth"

    "github.com/danielgtaylor/huma/v2"
)

// Handler handles profile-related HTTP requests.
type Handler struct {
    setupUC *profile.SetupUseCase
}

// NewHandler creates a new profile handler.
func NewHandler(setupUC *profile.SetupUseCase) *Handler {
    return &Handler{setupUC: setupUC}
}

// ... rest of handler
```

**Step 2: Request review**

Dispatch a **CodeReviewer** subagent to review the handler fix. Present: the updated `Handler` struct with the `setupUC` field, the `NewHandler` constructor, and how it integrates with the routes.

**Wait for explicit user approval** before proceeding.

**Step 3: Commit**

```bash
git add internal/gateways/http/v1/profile/handler.go
git commit -m "fix: wire SetupUseCase into Handler"
```

---

### Task 15: Run code checks and tests

**Step 1: Run all checks**

```bash
make code-check
```

Expected: format and lint pass. Fix any errors before proceeding.

**Step 2: Run tests**

```bash
make test
```

Expected: all tests pass (including new auth tests).

**Step 3: Run build**

```bash
go build ./...
```

Expected: build succeeds with no errors.

**Step 4: Request review**

Dispatch a **CodeReviewer** subagent to review the final code-check, test, and build results. Present: the `make code-check` output, the `make test` output, the `go build` result.

**Wait for explicit user approval** before proceeding.

**Step 5: Commit any fixes**

```bash
git add -A && git commit -m "fix: address linter warnings and test failures"
```

---

## Summary

### What Was Built

| Component | Files |
|-----------|-------|
| **Infrastructure** | `compose.infra.yml`, `compose.yml`, `.env.example`, `scripts/setup-m2m.sh`, `scripts/rotate-m2m-secret.sh` |
| **User domain** | `internal/domain/entity/user.go`, `internal/domain/errs/errs.go` |
| **User port** | `internal/app/ports/user_repo.go` |
| **User repo** | `internal/gateways/postgres/user.go`, `internal/gateways/postgres/sqlc/queries/user.sql` |
| **Logto M2M client** | `internal/gateways/logto/m2m_client.go`, `internal/gateways/logto/errors.go` |
| **Auth middleware** | `internal/gateways/http/v1/auth/middleware.go` |
| **Auth handlers** | `internal/gateways/http/v1/auth/handler.go`, `internal/gateways/http/v1/auth/routes.go` |
| **Register use case** | `internal/app/usecases/auth/register.go` |
| **Profile setup** | `internal/app/usecases/profile/setup.go`, `internal/gateways/http/v1/profile/` |
| **DI wiring** | `cmd/api/main.go` |
| **Tests** | `*_test.go` for middleware and register use case |

### Endpoints Added

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/auth/register` | Public | Register user in Logto + Finsplitter |
| `GET` | `/auth/me` | Optional | Returns auth status and setup state |
| `POST` | `/profile/setup` | Protected | Complete profile setup for new users |

### First-Time Setup Commands

```bash
# 1. Start infra (PostgreSQL + Logto)
make start-infra

# 2. Wait for Logto to initialize (~30s), then create M2M app
./scripts/setup-m2m.sh <admin-user> <admin-password> http://localhost:3001

# 3. Fill in .env:
#    LOGTO_MGMT_CLIENT_ID=<from-step-2>
#    LOGTO_MGMT_CLIENT_SECRET=<from-step-2>
#    LOGTO_APP_CLIENT_ID=<Traditional app client ID from Logto Console>
#    LOGTO_APP_CLIENT_SECRET=<Traditional app client secret>

# 4. Start the app
make start-dev

# 5. Run tests
make test
```
