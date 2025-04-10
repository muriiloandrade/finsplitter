-- Migration: Cria a tabela 'card_brand' e seus índices, se não existirem

CREATE TABLE IF NOT EXISTS "card_brand" (
    "id" uuid NOT NULL DEFAULT uuid_generate_v4(),
    "name" varchar NOT NULL,
    "created_date" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "last_modified_date" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, -- Será atualizado por trigger
    PRIMARY KEY ("id")
);

-- Índice único para garantir nomes de bandeiras únicos (case-insensitive).
CREATE UNIQUE INDEX IF NOT EXISTS "card_brand_unique_lower_name" ON "card_brand" (lower("name"));