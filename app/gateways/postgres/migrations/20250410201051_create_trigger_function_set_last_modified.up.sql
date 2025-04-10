-- Migration: Cria a função de trigger para atualizar last_modified_date.
-- Esta função será usada por triggers BEFORE UPDATE.

CREATE OR REPLACE FUNCTION trigger_set_last_modified()
RETURNS TRIGGER AS $$
BEGIN
  -- Define last_modified_date para o tempo atual no registro que está sendo atualizado.
  NEW.last_modified_date = NOW(); -- ou CURRENT_TIMESTAMP
  RETURN NEW; -- Retorna o registro modificado para operações BEFORE trigger.
END;
$$ LANGUAGE plpgsql;