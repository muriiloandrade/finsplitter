-- Migration: Removes value validation CHECK constraints

ALTER TABLE "transaction"
DROP CONSTRAINT IF EXISTS "check_transaction_value_not_zero";

ALTER TABLE "transaction"
DROP CONSTRAINT IF EXISTS "check_transaction_installments_positive";
