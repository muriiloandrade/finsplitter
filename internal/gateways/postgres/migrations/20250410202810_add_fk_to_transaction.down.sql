-- Down Migration: Remove the foreign key constraints from 'transaction' referencing 'card' and 'bill'.

ALTER TABLE "transaction"
DROP CONSTRAINT IF EXISTS "fk_transaction_card_id_card_id",
DROP CONSTRAINT IF EXISTS "fk_transaction_bill_id_bill_id";