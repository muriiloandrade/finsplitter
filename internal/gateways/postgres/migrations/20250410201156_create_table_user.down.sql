-- Down Migration: Drop the 'user' table and its associated indexes.

DROP INDEX IF EXISTS "user_unique_lower_email";
DROP INDEX IF EXISTS "user_unique_lower_username";
DROP TABLE IF EXISTS "user";