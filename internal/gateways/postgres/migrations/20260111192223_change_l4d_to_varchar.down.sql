-- Migration: Reverts card.l4d from VARCHAR(4) back to smallint

-- Drop the CHECK constraint
ALTER TABLE "card"
DROP CONSTRAINT IF EXISTS "check_card_l4d_format";

-- Convert VARCHAR back to smallint (this will remove leading zeros)
ALTER TABLE "card"
ALTER COLUMN "l4d" TYPE smallint 
USING CASE 
    WHEN "l4d" IS NULL THEN NULL
    ELSE "l4d"::smallint
END;
