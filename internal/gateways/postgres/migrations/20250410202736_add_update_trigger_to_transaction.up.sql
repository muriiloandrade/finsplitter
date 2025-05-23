-- Migration: Adiciona o trigger de atualização de timestamp para a tabela 'transaction'.

CREATE TRIGGER trigger_transaction_set_last_modified
BEFORE UPDATE ON "transaction"
FOR EACH ROW
EXECUTE FUNCTION trigger_set_last_modified();