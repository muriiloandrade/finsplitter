-- Down Migration: Drop the update timestamp trigger from the 'user' table.
DROP TRIGGER IF EXISTS trigger_user_set_last_modified ON "user";