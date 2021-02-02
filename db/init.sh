#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE DATABASE blog_dev;
    CREATE DATABASE blog_test;
    GRANT ALL PRIVILEGES ON DATABASE blog_dev TO goblog;
    \c blog_dev
    \i /init_schema/schema.sql
EOSQL
