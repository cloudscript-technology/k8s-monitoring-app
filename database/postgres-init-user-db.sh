#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    -- Cria o usuário 'api' se não existir
    DO \$\$
    BEGIN
        IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'api') THEN
            CREATE ROLE api WITH SUPERUSER LOGIN PASSWORD 'api';
        END IF;
    END
    \$\$;

    -- Cria o banco de dados 'api' se não existir
    DO \$\$
    BEGIN
        IF NOT EXISTS (SELECT FROM pg_database WHERE datname = 'api') THEN
            CREATE DATABASE api;
        END IF;
    END
    \$\$;

    -- Concede todos os privilégios no banco de dados 'api' para o usuário 'api'
    GRANT ALL PRIVILEGES ON DATABASE api TO api;
EOSQL