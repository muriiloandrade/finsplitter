-- Migration: Adiciona o trigger de atualização de timestamp para a tabela 'person'.

CREATE TRIGGER trigger_person_set_last_modified
BEFORE UPDATE ON "person"
FOR EACH ROW
EXECUTE FUNCTION trigger_set_last_modified();