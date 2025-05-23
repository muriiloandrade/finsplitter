-- Migration: Habilita a extensão uuid-ossp se não estiver habilitada.
-- Necessária para a função uuid_generate_v4().

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";