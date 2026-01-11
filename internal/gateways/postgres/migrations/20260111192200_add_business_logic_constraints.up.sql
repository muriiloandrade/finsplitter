-- Migration: Adds CHECK constraints to enforce business logic rules

-- Bill paid_on can only be set when paid is true
-- If paid is false, paid_on must be NULL
-- If paid is true, paid_on can be NULL or have a date (optional strict mode would require NOT NULL)
ALTER TABLE "bill"
ADD CONSTRAINT "check_bill_paid_on_logic" CHECK (
    ("paid" = false AND "paid_on" IS NULL) OR 
    ("paid" = true)
);

-- Note: The percentage constraints (0-100) already exist in the schema:
-- - check_card_person_percentage (added in migration 20250410203349)
-- - check_transaction_person_percentage (added in migration 20250410203703)
