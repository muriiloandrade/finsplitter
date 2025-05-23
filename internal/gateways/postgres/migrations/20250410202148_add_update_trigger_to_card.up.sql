-- Migration: Adiciona o trigger de atualização de timestamp para a tabela 'card'.

CREATE TRIGGER trigger_card_set_last_modified
BEFORE UPDATE ON "card"
FOR EACH ROW
EXECUTE FUNCTION trigger_set_last_modified();