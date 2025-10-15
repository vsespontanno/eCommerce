-- +goose Up
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    productID BIGINT NOT NULL UNIQUE,
    productName VARCHAR(255),
    productDescription VARCHAR(255),
    productPrice INT NOT NULL DEFAULT 0,
    productQuantity INT NOT NULL DEFAULT 0,
    reserved INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_id ON products (id);
CREATE INDEX IF NOT EXISTS idx_product_id ON products (productID);
-- +goose Down
DROP TABLE IF EXISTS products;