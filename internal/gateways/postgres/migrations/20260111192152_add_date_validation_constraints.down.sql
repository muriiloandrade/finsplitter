-- Migration: Removes date validation CHECK constraints

ALTER TABLE "bill"
DROP CONSTRAINT IF EXISTS "check_bill_month_range";

ALTER TABLE "bill"
DROP CONSTRAINT IF EXISTS "check_bill_due_month_range";

ALTER TABLE "bill"
DROP CONSTRAINT IF EXISTS "check_bill_year_range";

ALTER TABLE "card"
DROP CONSTRAINT IF EXISTS "check_card_due_date_range";

ALTER TABLE "card"
DROP CONSTRAINT IF EXISTS "check_card_closing_date_range";
