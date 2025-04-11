-- Down Migration: Drop the update timestamp trigger from the 'person' table.

DROP TRIGGER IF EXISTS trigger_person_set_last_modified ON "person";