-- Migration: Reverts decimal columns back to unlimited precision

-- Revert transaction values back to unspecified decimal
ALTER TABLE "transaction"
ALTER COLUMN "value" TYPE DECIMAL;

-- Revert transaction person calculated values
ALTER TABLE "transaction_person"
ALTER COLUMN "calculated_value" TYPE DECIMAL;

-- Revert percentage columns
ALTER TABLE "card_person"
ALTER COLUMN "default_percentage" TYPE DECIMAL;

ALTER TABLE "transaction_person"
ALTER COLUMN "percentage" TYPE DECIMAL;
