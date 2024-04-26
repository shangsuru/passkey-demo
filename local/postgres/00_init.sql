CREATE DATABASE mydb;

\c mydb;

CREATE ROLE myuser
WITH
  LOGIN PASSWORD 'mypassword';

GRANT ALL PRIVILEGES ON SCHEMA public TO myuser;
