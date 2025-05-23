-- Migration: Adiciona o trigger de atualização de timestamp para a tabela 'bill'.

CREATE TRIGGER trigger_bill_set_last_modified
BEFORE UPDATE ON "bill"
FOR EACH ROW
EXECUTE FUNCTION trigger_set_last_modified();