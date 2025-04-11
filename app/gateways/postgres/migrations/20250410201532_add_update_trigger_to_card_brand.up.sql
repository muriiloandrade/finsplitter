-- Migration: Adiciona o trigger de atualização de timestamp para a tabela 'card_brand'.

CREATE TRIGGER trigger_card_brand_set_last_modified
BEFORE UPDATE ON "card_brand"
FOR EACH ROW
EXECUTE FUNCTION trigger_set_last_modified();