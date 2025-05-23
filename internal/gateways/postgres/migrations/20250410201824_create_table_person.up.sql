-- Migration: Cria a tabela 'person' se ela não existir

CREATE TABLE IF NOT EXISTS "person" (
    "id" uuid NOT NULL DEFAULT uuid_generate_v4(),
    "user_id" uuid, -- FK para 'user' será adicionada posteriormente
    "name" varchar NOT NULL,
    "created_date" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "last_modified_date" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, -- Será atualizado por trigger
    PRIMARY KEY ("id")
);

CREATE INDEX IF NOT EXISTS "idx_person_user_id" ON "person" ("user_id");