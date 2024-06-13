#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
	-- Create test database
  CREATE DATABASE testdb;

  -- Create client user
  CREATE USER "postgres-user" WITH ENCRYPTED PASSWORD '1234';

  -- Grant privileges to postgres-user
  GRANT ALL PRIVILEGES ON DATABASE testdb TO "postgres-user";
EOSQL
