
-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id         SERIAL PRIMARY KEY,
    userID     BIGINT   NOT NULL UNIQUE,
    email      TEXT   NOT NULL UNIQUE,
    first_name TEXT  NOT NULL,
    last_name  TEXT  NOT NULL,
    pass_hash BYTEA  NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_email ON users (email);

-- +goose Down
DROP INDEX IF EXISTS idx_email;
DROP TABLE IF EXISTS users;

