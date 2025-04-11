-- Down Migration: Remove the foreign key constraint from 'person' referencing 'user'.

ALTER TABLE "person"
DROP CONSTRAINT IF EXISTS "fk_person_user_id_user_id";