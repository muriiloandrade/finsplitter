-- Migration: Cria a tabela de junção 'card_person' se ela não existir

CREATE TABLE IF NOT EXISTS "card_person" (
    "card_id" uuid NOT NULL, -- FK para 'card' será adicionada posteriormente
    "person_id" uuid NOT NULL, -- FK para 'person' será adicionada posteriormente
    "default_percentage" decimal,
    "created_date" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "end_date" timestamp
);

CREATE INDEX IF NOT EXISTS "idx_card_person_card_id" ON "card_person" ("card_id");
CREATE INDEX IF NOT EXISTS "idx_card_person_person_id" ON "card_person" ("person_id");