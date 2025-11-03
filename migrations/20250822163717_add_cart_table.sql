-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS cart (
    id SERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    product_id BIGINT NOT NULL REFERENCES products(id),
    quantity INT NOT NULL,
    amount_for_product INT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_id ON cart (id);
CREATE INDEX IF NOT EXISTS idx_user_id ON cart (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS cart;
-- +goose StatementEnd
