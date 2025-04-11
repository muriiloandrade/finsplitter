-- Migration: Cria a tabela 'card' se ela não existir e adiciona comentários

CREATE TABLE IF NOT EXISTS "card" (
    "id" uuid NOT NULL DEFAULT uuid_generate_v4(),
    "brand_id" uuid, -- FK para 'card_brand' será adicionada posteriormente
    "user_id" uuid NOT NULL, -- FK para 'user' será adicionada posteriormente
    "name" varchar NOT NULL,
    "l4d" smallint,
    "due_date" smallint NOT NULL,
    "closing_date" smallint,
    "tier" varchar,
    "created_date" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "last_modified_date" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, -- Será atualizado por trigger
    PRIMARY KEY ("id")
);

-- Comentários...
COMMENT ON COLUMN "card"."l4d" IS 'Last 4 digits, corresponde aos 4 últimos dígitos do número docartão';
COMMENT ON COLUMN "card"."due_date" IS 'O dia do mês que a fatura vence';
COMMENT ON COLUMN "card"."closing_date" IS 'O dia do mês que a fatura fecha';
COMMENT ON COLUMN "card"."tier" IS 'Exemplos: Black, Platinum, entre outros';

CREATE INDEX IF NOT EXISTS "idx_card_brand_id" ON "card" ("brand_id");
CREATE INDEX IF NOT EXISTS "idx_card_user_id" ON "card" ("user_id");