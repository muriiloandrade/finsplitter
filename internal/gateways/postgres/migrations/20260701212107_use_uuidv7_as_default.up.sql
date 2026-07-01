-- Migration: Replace uuid_generate_v4() default with uuidv7() (built-in PG18)
-- uuidv7 generates time-ordered UUIDs, improving index performance
-- and letting us drop the uuid-ossp extension.

ALTER TABLE "card_brand" ALTER COLUMN "id" SET DEFAULT uuidv7();
ALTER TABLE "card" ALTER COLUMN "id" SET DEFAULT uuidv7();
ALTER TABLE "person" ALTER COLUMN "id" SET DEFAULT uuidv7();
ALTER TABLE "transaction" ALTER COLUMN "id" SET DEFAULT uuidv7();
ALTER TABLE "bill" ALTER COLUMN "id" SET DEFAULT uuidv7();
ALTER TABLE "user" ALTER COLUMN "id" SET DEFAULT uuidv7();

-- uuidv7() is built into PostgreSQL 18 -- no extension needed.
DROP EXTENSION IF EXISTS "uuid-ossp";
