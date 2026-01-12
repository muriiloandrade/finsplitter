-- Migration: Adds reasonable length limits to VARCHAR columns to prevent abuse

-- User table
ALTER TABLE "user" ALTER COLUMN "name" TYPE VARCHAR(255);
ALTER TABLE "user" ALTER COLUMN "email" TYPE VARCHAR(255);
ALTER TABLE "user" ALTER COLUMN "username" TYPE VARCHAR(100);
ALTER TABLE "user" ALTER COLUMN "phone_number" TYPE VARCHAR(20);
ALTER TABLE "user" ALTER COLUMN "password_hash" TYPE VARCHAR(255);

-- Card brand table
ALTER TABLE "card_brand" ALTER COLUMN "name" TYPE VARCHAR(100);

-- Person table
ALTER TABLE "person" ALTER COLUMN "name" TYPE VARCHAR(255);

-- Card table
ALTER TABLE "card" ALTER COLUMN "name" TYPE VARCHAR(100);
ALTER TABLE "card" ALTER COLUMN "tier" TYPE VARCHAR(50);

-- Transaction table
ALTER TABLE "transaction" ALTER COLUMN "identifier" TYPE VARCHAR(255);
ALTER TABLE "transaction" ALTER COLUMN "name" TYPE VARCHAR(255);
