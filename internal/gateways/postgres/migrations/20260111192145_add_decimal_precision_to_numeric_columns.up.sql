-- Migration: Adds proper precision to decimal/numeric columns for financial accuracy
-- NUMERIC(15, 2) allows values up to 999,999,999,999.99 (15 total digits, 2 after decimal)
-- NUMERIC(5, 2) allows values from 0.00 to 100.00 (for percentages)

-- Transaction values (monetary amounts)
ALTER TABLE "transaction"
ALTER COLUMN "value" TYPE NUMERIC(15, 2);

-- Transaction person calculated values (monetary amounts)
ALTER TABLE "transaction_person"
ALTER COLUMN "calculated_value" TYPE NUMERIC(15, 2);

-- Percentage columns (0.00 to 100.00)
ALTER TABLE "card_person"
ALTER COLUMN "default_percentage" TYPE NUMERIC(5, 2);

ALTER TABLE "transaction_person"
ALTER COLUMN "percentage" TYPE NUMERIC(5, 2);
