-- Down Migration: Remove the foreign key constraints from 'card'.

ALTER TABLE "card"
DROP CONSTRAINT IF EXISTS "fk_card_user_id_user_id",
DROP CONSTRAINT IF EXISTS "fk_card_brand_id_card_brand_id";