-- +goose Up
CREATE TABLE IF NOT EXISTS cart (
    id SERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    product_id BIGINT NOT NULL REFERENCES products(id),
    quantity INT NOT NULL,
    amount_for_product INT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_cart_user_product ON cart (user_id, product_id);
CREATE INDEX IF NOT EXISTS idx_cart_user_id ON cart (user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_cart_user_id;
DROP INDEX IF EXISTS idx_cart_user_product;
DROP TABLE IF EXISTS cart;