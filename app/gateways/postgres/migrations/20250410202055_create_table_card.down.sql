-- Down Migration: Drop the 'card' table and its indexes.

DROP INDEX IF EXISTS "idx_card_user_id";
DROP INDEX IF EXISTS "idx_card_brand_id";
DROP TABLE IF EXISTS "card";