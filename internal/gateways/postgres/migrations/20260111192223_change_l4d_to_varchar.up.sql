-- Migration: Changes card.l4d from smallint to VARCHAR(4) to preserve leading zeros
-- Example: "0000", "0001", "0123", "9999"

-- Convert existing smallint values to VARCHAR with zero-padding
ALTER TABLE "card" 
ALTER COLUMN "l4d" TYPE VARCHAR(4) 
USING CASE 
    WHEN "l4d" IS NULL THEN NULL
    ELSE LPAD("l4d"::text, 4, '0')
END;

-- Add CHECK constraint to ensure it's exactly 4 numeric digits
ALTER TABLE "card"
ADD CONSTRAINT "check_card_l4d_format" CHECK ("l4d" IS NULL OR "l4d" ~ '^[0-9]{4}$');
