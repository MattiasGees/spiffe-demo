#!/bin/bash
set -e

cp /pg_hba.conf /var/lib/postgresql/data/pg_hba.conf
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
  SELECT pg_reload_conf();
EOSQL
