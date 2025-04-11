-- Migration: Cria a tabela 'user' se ela não existir
-- Tabela base sem dependências externas diretas no schema

CREATE TABLE IF NOT EXISTS "user" (
    "id" uuid NOT NULL DEFAULT uuid_generate_v4(),
    "name" varchar NOT NULL,
    "email" varchar NOT NULL,
    "phone_number" varchar,
    "username" varchar NOT NULL,
    "password_hash" varchar NOT NULL,
    "created_date" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "last_modified_date" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, -- Será atualizado por trigger
    PRIMARY KEY ("id")
);

CREATE UNIQUE INDEX IF NOT EXISTS "user_unique_lower_username" ON "user" (lower("username"));
CREATE UNIQUE INDEX IF NOT EXISTS "user_unique_lower_email" ON "user" (lower("email"));