-- Migration: Revert uuidv7() default back to uuid_generate_v4().

-- Re-enable the uuid-ossp extension before restoring the old defaults.
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

ALTER TABLE "card_brand" ALTER COLUMN "id" SET DEFAULT uuid_generate_v4();
ALTER TABLE "card" ALTER COLUMN "id" SET DEFAULT uuid_generate_v4();
ALTER TABLE "person" ALTER COLUMN "id" SET DEFAULT uuid_generate_v4();
ALTER TABLE "transaction" ALTER COLUMN "id" SET DEFAULT uuid_generate_v4();
ALTER TABLE "bill" ALTER COLUMN "id" SET DEFAULT uuid_generate_v4();
ALTER TABLE "user" ALTER COLUMN "id" SET DEFAULT uuid_generate_v4();
