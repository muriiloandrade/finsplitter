-- Down Migration: Drop the 'card_brand' table and its index.

DROP INDEX IF EXISTS "card_brand_unique_lower_name";
DROP TABLE IF EXISTS "card_brand";