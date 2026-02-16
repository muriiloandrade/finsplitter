<!-- Context: project-intelligence/database | Priority: high | Version: 1.0 | Updated: 2026-02-16 -->

# Database Patterns

**Purpose**: PostgreSQL and database development patterns for finsplitter.
**Last Updated**: 2026-02-16

## Naming Conventions
- **Tables**: singular, snake_case (`customer`, `card_brand`)
- **Columns**: snake_case (`user_id`, `created_at`)
- **Functions**: snake_case, verb prefix (`create_card_brand`, `get_user_by_id`)

## SQLC Query Pattern
```sql
-- name: CreateCardBrand :one
INSERT INTO card_brand (name) VALUES ($1) RETURNING *;

-- name: GetCardBrandByID :one
SELECT * FROM card_brand WHERE id = $1;
```

## Constraints
- Always use `NOT NULL` where applicable
- Use `UNIQUE` for columns requiring uniqueness
- Use `CHECK` for business rules
- Use `FOREIGN KEY` for relationships

## Indexing
```sql
CREATE INDEX idx_card_brand_name ON card_brand(name);
-- Partial index
CREATE INDEX idx_active_users ON users(id) WHERE status = 'active';
```

## Transactions
```sql
BEGIN;
-- Operations
COMMIT;
-- Or on error
ROLLBACK;
```

## Security
- Parameterized queries (sqlc generates)
- Use `quote_literal()` for dynamic SQL
- Grant minimum privileges
- Use roles for permission management

## 📂 Codebase References
- **Queries**: `internal/gateways/postgres/sqlc/queries/`
- **Migrations**: `internal/gateways/postgres/migrations/`
- **Repo Impl**: `internal/gateways/postgres/`
