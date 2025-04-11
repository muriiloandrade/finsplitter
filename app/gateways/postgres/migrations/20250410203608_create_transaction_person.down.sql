-- Down Migration: Drop the 'transaction_person' table and its indexes.

DROP INDEX IF EXISTS "idx_transaction_person_person_id";
DROP INDEX IF EXISTS "idx_transaction_person_transaction_id";
DROP TABLE IF EXISTS "transaction_person";