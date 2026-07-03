-- Migration: Remove PII columns from user table
-- Profile data (name, email, phone_number, username) now lives only in Logto.
-- The local user table acts as a lightweight link between Logto identity and
-- Finsplitter records.
--
-- Constraints to drop before columns:
--   user_unique_lower_username (unique index on lower(username))
--   user_unique_lower_email    (unique index on lower(email))
--   user_pkey                  (id is still the PK, stays)

DROP INDEX IF EXISTS "user_unique_lower_username";
DROP INDEX IF EXISTS "user_unique_lower_email";

ALTER TABLE "user"
    DROP COLUMN IF EXISTS "name",
    DROP COLUMN IF EXISTS "email",
    DROP COLUMN IF EXISTS "phone_number",
    DROP COLUMN IF EXISTS "username",
    DROP COLUMN IF EXISTS "password_hash";
