#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
	-- Create test database
  CREATE DATABASE testdb;

  -- Create client user
  CREATE USER "spiffe-customer" WITH ENCRYPTED PASSWORD '1234';

  -- Grant privileges to spiffe-customer
  GRANT ALL PRIVILEGES ON DATABASE testdb TO "spiffe-customer";

  -- Connect to the newly created database
  \c testdb;

  -- Create the table with the specified columns
  CREATE TABLE test_table (
      name VARCHAR(255),
      text VARCHAR(255)
  );
EOSQL
