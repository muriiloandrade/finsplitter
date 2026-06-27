-- Migration: Remove logto_user_id column from user table

DROP INDEX IF EXISTS "user_unique_logto_user_id";

ALTER TABLE "user"
    DROP COLUMN IF EXISTS "logto_user_id",
    ALTER COLUMN "password_hash" SET NOT NULL;
