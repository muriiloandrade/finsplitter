-- Migration: Adiciona chaves estrangeiras PARA a tabela 'card'

ALTER TABLE "card"
ADD CONSTRAINT "fk_card_brand_id_card_brand_id" FOREIGN KEY("brand_id") REFERENCES "card_brand"("id"),
ADD CONSTRAINT "fk_card_user_id_user_id" FOREIGN KEY("user_id") REFERENCES "user"("id");