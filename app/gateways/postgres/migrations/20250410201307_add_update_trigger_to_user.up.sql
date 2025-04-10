-- Migration: Adiciona o trigger de atualização de timestamp para a tabela 'user'.

-- Cria o trigger para executar a função ANTES de cada UPDATE na tabela 'user'
CREATE TRIGGER trigger_user_set_last_modified
BEFORE UPDATE ON "user"
FOR EACH ROW
EXECUTE FUNCTION trigger_set_last_modified();