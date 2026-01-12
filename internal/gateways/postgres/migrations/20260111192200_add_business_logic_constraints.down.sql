-- Migration: Removes business logic CHECK constraints

ALTER TABLE "bill"
DROP CONSTRAINT IF EXISTS "check_bill_paid_on_logic";
