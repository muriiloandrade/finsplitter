-- Migration: Cria a tabela de junção 'transaction_person' se ela não existir

CREATE TABLE IF NOT EXISTS "transaction_person" (
    "person_id" uuid NOT NULL, -- FK para 'person' será adicionada posteriormente
    "transaction_id" uuid NOT NULL, -- FK para 'transaction' será adicionada posteriormente
    "percentage" decimal,
    "calculated_value" decimal NOT NULL,
    "created_date" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "end_date" timestamp
);

CREATE INDEX IF NOT EXISTS "idx_transaction_person_person_id" ON "transaction_person" ("person_id");
CREATE INDEX IF NOT EXISTS "idx_transaction_person_transaction_id" ON "transaction_person" ("transaction_id");