
-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id        SERIAL PRIMARY KEY,
    email     TEXT   NOT NULL UNIQUE,
    first_name TEXT  NOT NULL,
    last_name  TEXT  NOT NULL,
    pass_hash BYTEA  NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_email ON users (email);

CREATE TABLE IF NOT EXISTS apps (
    id     SERIAL PRIMARY KEY,
    name   TEXT NOT NULL UNIQUE,
    secret TEXT NOT NULL UNIQUE
);

-- +goose Down
DROP TABLE IF EXISTS apps;
DROP INDEX IF EXISTS idx_email;
DROP TABLE IF EXISTS users;

