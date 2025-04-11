-- Migration: Adiciona chaves estrangeiras PARA a tabela 'bill'

ALTER TABLE "bill"
ADD CONSTRAINT "fk_bill_card_id_card_id" FOREIGN KEY("card_id") REFERENCES "card"("id");