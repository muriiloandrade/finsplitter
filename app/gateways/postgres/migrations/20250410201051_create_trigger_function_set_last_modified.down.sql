-- Down Migration: Drop the trigger function used for updating last_modified_date.
-- Note: This will fail if any triggers still use this function.

DROP FUNCTION IF EXISTS trigger_set_last_modified();