-- Migration: Cria a tabela 'bill' se ela não existir
-- Nota: Esta tabela não possui 'last_modified_date'.

CREATE TABLE IF NOT EXISTS "bill" (
    "id" uuid NOT NULL DEFAULT uuid_generate_v4(),
    "card_id" uuid, -- FK para 'card' será adicionada posteriormente
    "month" smallint NOT NULL,
    "year" smallint NOT NULL,
    "due_date" smallint NOT NULL,
    "due_month" smallint NOT NULL,
    "paid" boolean DEFAULT false,
    "paid_on" date,
    "created_date" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "last_modified_date" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, -- Será atualizado por trigger
    PRIMARY KEY ("id")
);

CREATE UNIQUE INDEX IF NOT EXISTS "bill_unique_card_month_year" ON "bill" ("card_id", "year", "month");