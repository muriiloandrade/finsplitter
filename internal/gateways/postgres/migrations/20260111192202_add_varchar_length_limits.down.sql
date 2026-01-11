-- Migration: Reverts VARCHAR columns back to unlimited length

-- User table
ALTER TABLE "user" ALTER COLUMN "name" TYPE VARCHAR;
ALTER TABLE "user" ALTER COLUMN "email" TYPE VARCHAR;
ALTER TABLE "user" ALTER COLUMN "username" TYPE VARCHAR;
ALTER TABLE "user" ALTER COLUMN "phone_number" TYPE VARCHAR;
ALTER TABLE "user" ALTER COLUMN "password_hash" TYPE VARCHAR;

-- Card brand table
ALTER TABLE "card_brand" ALTER COLUMN "name" TYPE VARCHAR;

-- Person table
ALTER TABLE "person" ALTER COLUMN "name" TYPE VARCHAR;

-- Card table
ALTER TABLE "card" ALTER COLUMN "name" TYPE VARCHAR;
ALTER TABLE "card" ALTER COLUMN "tier" TYPE VARCHAR;

-- Transaction table
ALTER TABLE "transaction" ALTER COLUMN "identifier" TYPE VARCHAR;
ALTER TABLE "transaction" ALTER COLUMN "name" TYPE VARCHAR;
