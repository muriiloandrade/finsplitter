-- Migration: Adiciona chaves estrangeiras PARA a tabela 'transaction_person'

ALTER TABLE "transaction_person"
ADD CONSTRAINT "pk_transaction_person" PRIMARY KEY ("transaction_id", "person_id"),
ADD CONSTRAINT "fk_transaction_person_transaction_id_transaction_id" FOREIGN KEY("transaction_id") REFERENCES "transaction"("id"),
ADD CONSTRAINT "fk_transaction_person_person_id_person_id" FOREIGN KEY("person_id") REFERENCES "person"("id"),
ADD CONSTRAINT "check_transaction_person_percentage" CHECK ("percentage" >= 0 AND "percentage" <= 100);