
-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id         BIGSERIAL PRIMARY KEY,
    public_id  UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    email      TEXT NOT NULL UNIQUE,
    first_name TEXT NOT NULL,
    last_name  TEXT NOT NULL,
    pass_hash  BYTEA NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_email ON users (email);
CREATE INDEX IF NOT EXISTS idx_public_id ON users (public_id);

-- +goose Down
DROP INDEX IF EXISTS idx_email;
DROP TABLE IF EXISTS users;

