-- Migration: Cria a tabela 'transaction' se ela não existir e adiciona comentários

CREATE TABLE IF NOT EXISTS "transaction" (
    "id" uuid NOT NULL DEFAULT uuid_generate_v4(),
    "card_id" uuid, -- FK para 'card' será adicionada posteriormente
    "bill_id" uuid, -- FK para 'bill' será adicionada posteriormente (pode ser NULL)
    "identifier" varchar,
    "name" varchar,
    "value" decimal NOT NULL,
    "date" date NOT NULL,
    "recurring_charge" boolean NOT NULL DEFAULT false,
    "installments_number" smallint,
    "created_date" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "last_modified_date" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, -- Será atualizado por trigger
    PRIMARY KEY ("id")
);

-- Comentários...
COMMENT ON COLUMN "transaction"."identifier" IS 'Identificador da transação - como consta na fatura do banco';
COMMENT ON COLUMN "transaction"."name" IS 'Como quer identificar essa transação';
COMMENT ON COLUMN "transaction"."recurring_charge" IS 'Indica se a transação estará presente em mais de 1 fatura';

CREATE INDEX IF NOT EXISTS "idx_transaction_card_id" ON "transaction" ("card_id");
CREATE INDEX IF NOT EXISTS "idx_transaction_bill_id" ON "transaction" ("bill_id");