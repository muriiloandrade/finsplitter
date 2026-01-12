-- Migration: Adds CHECK constraints to validate financial values and prevent invalid data
-- Note: Transaction values can be positive (debits/purchases) or negative (credits/refunds)

-- Transaction value must not be zero (can be positive for debits or negative for credits)
ALTER TABLE "transaction"
ADD CONSTRAINT "check_transaction_value_not_zero" CHECK ("value" != 0);

-- Transaction person calculated value can be positive or negative (no constraint needed for sign)
-- We just ensure it exists (NOT NULL is already enforced by the column definition)

-- Transaction installments must be positive if specified
ALTER TABLE "transaction"
ADD CONSTRAINT "check_transaction_installments_positive" CHECK ("installments_number" IS NULL OR "installments_number" > 0);
