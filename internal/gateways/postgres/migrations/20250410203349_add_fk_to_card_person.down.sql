-- Down Migration: Remove the foreign key constraints from 'card_person' referencing 'card' and 'person'.

ALTER TABLE "card_person"
DROP CONSTRAINT IF EXISTS "check_card_person_percentage",
DROP CONSTRAINT IF EXISTS "fk_card_person_person_id_person_id",
DROP CONSTRAINT IF EXISTS "fk_card_person_card_id_card_id",
DROP CONSTRAINT IF EXISTS "pk_card_person";
