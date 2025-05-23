-- Down Migration: Drop the 'card_person' table and its indexes.

DROP INDEX IF EXISTS "idx_card_person_person_id";
DROP INDEX IF EXISTS "idx_card_person_card_id";
DROP TABLE IF EXISTS "card_person";