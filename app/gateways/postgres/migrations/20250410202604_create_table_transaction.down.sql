-- Down Migration: Drop the 'transaction' table and its indexes.

DROP INDEX IF EXISTS "idx_transaction_bill_id";
DROP INDEX IF EXISTS "idx_transaction_card_id";
DROP TABLE IF EXISTS "transaction";