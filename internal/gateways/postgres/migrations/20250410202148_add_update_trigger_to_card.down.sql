-- Down Migration: Drop the update timestamp trigger from the 'card' table.

DROP TRIGGER IF EXISTS trigger_card_set_last_modified ON "card";