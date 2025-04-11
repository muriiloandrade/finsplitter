-- Down Migration: Remove the foreign key constraints from 'transaction_person' referencing 'transaction' and 'person'.

ALTER TABLE "transaction_person"
DROP CONSTRAINT IF EXISTS "check_transaction_person_percentage",
DROP CONSTRAINT IF EXISTS "fk_transaction_person_person_id_person_id",
DROP CONSTRAINT IF EXISTS "fk_transaction_person_transaction_id_transaction_id",
DROP CONSTRAINT IF EXISTS "pk_transaction_person";
