-- Migration: Adiciona chaves estrangeiras PARA a tabela 'card_person'

ALTER TABLE "card_person"
ADD CONSTRAINT "pk_card_person" PRIMARY KEY ("card_id", "person_id"),
ADD CONSTRAINT "fk_card_person_card_id_card_id" FOREIGN KEY("card_id") REFERENCES "card"("id"),
ADD CONSTRAINT "fk_card_person_person_id_person_id" FOREIGN KEY("person_id") REFERENCES "person"("id"),
ADD CONSTRAINT "check_card_person_percentage" CHECK ("default_percentage" >= 0 AND "default_percentage" <= 100);