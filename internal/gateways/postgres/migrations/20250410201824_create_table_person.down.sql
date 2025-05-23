-- Down Migration: Drop the 'person' table and its index.

DROP INDEX IF EXISTS "idx_person_user_id";
DROP TABLE IF EXISTS "person";