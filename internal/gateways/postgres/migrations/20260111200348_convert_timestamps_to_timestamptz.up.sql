-- Migration: Converts all TIMESTAMP columns to TIMESTAMPTZ (timestamp with time zone)
-- This ensures consistent timezone handling across different environments and locations.
-- PostgreSQL automatically converts existing timestamps to UTC during the migration.

-- User table
ALTER TABLE "user"
ALTER COLUMN "created_date" TYPE TIMESTAMPTZ USING "created_date" AT TIME ZONE 'UTC',
ALTER COLUMN "last_modified_date" TYPE TIMESTAMPTZ USING "last_modified_date" AT TIME ZONE 'UTC';

-- Card brand table
ALTER TABLE "card_brand"
ALTER COLUMN "created_date" TYPE TIMESTAMPTZ USING "created_date" AT TIME ZONE 'UTC',
ALTER COLUMN "last_modified_date" TYPE TIMESTAMPTZ USING "last_modified_date" AT TIME ZONE 'UTC';

-- Person table
ALTER TABLE "person"
ALTER COLUMN "created_date" TYPE TIMESTAMPTZ USING "created_date" AT TIME ZONE 'UTC',
ALTER COLUMN "last_modified_date" TYPE TIMESTAMPTZ USING "last_modified_date" AT TIME ZONE 'UTC';

-- Card table
ALTER TABLE "card"
ALTER COLUMN "created_date" TYPE TIMESTAMPTZ USING "created_date" AT TIME ZONE 'UTC',
ALTER COLUMN "last_modified_date" TYPE TIMESTAMPTZ USING "last_modified_date" AT TIME ZONE 'UTC';

-- Bill table
ALTER TABLE "bill"
ALTER COLUMN "created_date" TYPE TIMESTAMPTZ USING "created_date" AT TIME ZONE 'UTC',
ALTER COLUMN "last_modified_date" TYPE TIMESTAMPTZ USING "last_modified_date" AT TIME ZONE 'UTC';

-- Transaction table
ALTER TABLE "transaction"
ALTER COLUMN "created_date" TYPE TIMESTAMPTZ USING "created_date" AT TIME ZONE 'UTC',
ALTER COLUMN "last_modified_date" TYPE TIMESTAMPTZ USING "last_modified_date" AT TIME ZONE 'UTC';

-- Card person junction table
ALTER TABLE "card_person"
ALTER COLUMN "created_date" TYPE TIMESTAMPTZ USING "created_date" AT TIME ZONE 'UTC',
ALTER COLUMN "end_date" TYPE TIMESTAMPTZ USING "end_date" AT TIME ZONE 'UTC';

-- Transaction person junction table
ALTER TABLE "transaction_person"
ALTER COLUMN "created_date" TYPE TIMESTAMPTZ USING "created_date" AT TIME ZONE 'UTC',
ALTER COLUMN "end_date" TYPE TIMESTAMPTZ USING "end_date" AT TIME ZONE 'UTC';
