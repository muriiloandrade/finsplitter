-- Down Migration: Drop the update timestamp trigger from the 'card_brand' table.
DROP TRIGGER IF EXISTS trigger_card_brand_set_last_modified ON "card_brand";