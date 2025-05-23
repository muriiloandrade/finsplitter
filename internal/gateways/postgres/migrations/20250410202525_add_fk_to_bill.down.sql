-- Down Migration: Remove the foreign key constraint from 'bill' referencing 'card'.
ALTER TABLE "bill"
DROP CONSTRAINT IF EXISTS "fk_bill_card_id_card_id";