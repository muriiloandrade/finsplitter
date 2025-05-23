-- Down Migration: Drop the update timestamp trigger from the 'bill' table.

DROP TRIGGER IF EXISTS trigger_bill_set_last_modified ON "bill";