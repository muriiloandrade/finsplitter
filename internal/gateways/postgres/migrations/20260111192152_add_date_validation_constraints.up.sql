-- Migration: Adds CHECK constraints to validate date-related columns

-- Bill month validation (1-12)
ALTER TABLE "bill"
ADD CONSTRAINT "check_bill_month_range" CHECK ("month" >= 1 AND "month" <= 12);

-- Bill due_month validation (1-12)
ALTER TABLE "bill"
ADD CONSTRAINT "check_bill_due_month_range" CHECK ("due_month" >= 1 AND "due_month" <= 12);

-- Bill year validation (reasonable range 2000-2100)
ALTER TABLE "bill"
ADD CONSTRAINT "check_bill_year_range" CHECK ("year" >= 2000 AND "year" <= 2100);

-- Card due_date validation (day of month: 1-31)
ALTER TABLE "card"
ADD CONSTRAINT "check_card_due_date_range" CHECK ("due_date" >= 1 AND "due_date" <= 31);

-- Card closing_date validation (day of month: 1-31 or NULL)
ALTER TABLE "card"
ADD CONSTRAINT "check_card_closing_date_range" CHECK ("closing_date" IS NULL OR ("closing_date" >= 1 AND "closing_date" <= 31));
