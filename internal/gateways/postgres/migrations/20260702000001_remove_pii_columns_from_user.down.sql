-- Migration down: Restore PII columns to user table

ALTER TABLE "user"
    ADD COLUMN IF NOT EXISTS "name" varchar,
    ADD COLUMN IF NOT EXISTS "email" varchar,
    ADD COLUMN IF NOT EXISTS "phone_number" varchar,
    ADD COLUMN IF NOT EXISTS "username" varchar,
    ADD COLUMN IF NOT EXISTS "password_hash" varchar;

-- Set defaults for existing rows (migration revert, data may be lost).
UPDATE "user" SET
    "name" = 'restored',
    "email" = 'restored@example.com',
    "username" = 'restored_' || id::text,
    "password_hash" = 'restored'
WHERE "name" IS NULL;

-- Now make them NOT NULL.
ALTER TABLE "user"
    ALTER COLUMN "name" SET NOT NULL,
    ALTER COLUMN "email" SET NOT NULL,
    ALTER COLUMN "username" SET NOT NULL,
    ALTER COLUMN "password_hash" SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS "user_unique_lower_username" ON "user" (lower("username"));
CREATE UNIQUE INDEX IF NOT EXISTS "user_unique_lower_email" ON "user" (lower("email"));
