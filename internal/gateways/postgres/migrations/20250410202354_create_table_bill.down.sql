-- Down Migration: Drop the 'bill' table and its unique index.

DROP INDEX IF EXISTS "bill_unique_card_month_year";
DROP TABLE IF EXISTS "bill";