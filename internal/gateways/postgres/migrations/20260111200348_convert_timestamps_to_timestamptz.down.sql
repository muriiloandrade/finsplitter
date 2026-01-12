-- Migration: Reverts TIMESTAMPTZ columns back to TIMESTAMP (without timezone)
-- Warning: This will lose timezone information. Timestamps will be stored as local time.

-- User table
ALTER TABLE "user"
ALTER COLUMN "created_date" TYPE TIMESTAMP,
ALTER COLUMN "last_modified_date" TYPE TIMESTAMP;

-- Card brand table
ALTER TABLE "card_brand"
ALTER COLUMN "created_date" TYPE TIMESTAMP,
ALTER COLUMN "last_modified_date" TYPE TIMESTAMP;

-- Person table
ALTER TABLE "person"
ALTER COLUMN "created_date" TYPE TIMESTAMP,
ALTER COLUMN "last_modified_date" TYPE TIMESTAMP;

-- Card table
ALTER TABLE "card"
ALTER COLUMN "created_date" TYPE TIMESTAMP,
ALTER COLUMN "last_modified_date" TYPE TIMESTAMP;

-- Bill table
ALTER TABLE "bill"
ALTER COLUMN "created_date" TYPE TIMESTAMP,
ALTER COLUMN "last_modified_date" TYPE TIMESTAMP;

-- Transaction table
ALTER TABLE "transaction"
ALTER COLUMN "created_date" TYPE TIMESTAMP,
ALTER COLUMN "last_modified_date" TYPE TIMESTAMP;

-- Card person junction table
ALTER TABLE "card_person"
ALTER COLUMN "created_date" TYPE TIMESTAMP,
ALTER COLUMN "end_date" TYPE TIMESTAMP;

-- Transaction person junction table
ALTER TABLE "transaction_person"
ALTER COLUMN "created_date" TYPE TIMESTAMP,
ALTER COLUMN "end_date" TYPE TIMESTAMP;
