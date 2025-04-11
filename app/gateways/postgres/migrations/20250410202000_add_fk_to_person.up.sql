-- Migration: Adiciona chaves estrangeiras PARA a tabela 'person'

ALTER TABLE "person"
ADD CONSTRAINT "fk_person_user_id_user_id" FOREIGN KEY("user_id") REFERENCES "user"("id");