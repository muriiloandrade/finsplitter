-- Migration: Add logto_user_id column to user table
-- Links Finsplitter users to their Logto identity

ALTER TABLE "user"
    ADD COLUMN IF NOT EXISTS "logto_user_id" varchar,
    ALTER COLUMN "password_hash" DROP NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS "user_unique_logto_user_id" ON "user" ("logto_user_id");
