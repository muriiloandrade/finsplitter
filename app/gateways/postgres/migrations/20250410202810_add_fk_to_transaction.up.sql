-- Migration: Adiciona chaves estrangeiras PARA a tabela 'transaction'

ALTER TABLE "transaction"
ADD CONSTRAINT "fk_transaction_card_id_card_id" FOREIGN KEY("card_id") REFERENCES "card"("id"),
ADD CONSTRAINT "fk_transaction_bill_id_bill_id" FOREIGN KEY("bill_id") REFERENCES "bill"("id"); -- Permite NULL