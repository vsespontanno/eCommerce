-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS wallets (
  user_id BIGINT PRIMARY KEY REFERENCES users(id),
  balance NUMERIC(12,2) NOT NULL DEFAULT 0,
  reserved NUMERIC(12,2) NOT NULL DEFAULT 0
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS wallets;
-- +goose StatementEnd
