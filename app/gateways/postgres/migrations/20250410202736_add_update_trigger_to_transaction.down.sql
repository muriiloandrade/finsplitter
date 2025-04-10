-- Down Migration: Drop the update timestamp trigger from the 'transaction' table.

DROP TRIGGER IF EXISTS trigger_transaction_set_last_modified ON "transaction";