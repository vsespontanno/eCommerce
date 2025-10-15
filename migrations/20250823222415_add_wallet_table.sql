-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS wallets (
  user_id BIGINT PRIMARY KEY REFERENCES users(id),
  balance INT NOT NULL DEFAULT 0,
  reserved INT NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_user_id ON wallets (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS wallets;
-- +goose StatementEnd
