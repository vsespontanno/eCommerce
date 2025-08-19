-- +goose Up
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    productID BIGINT NOT NULL UNIQUE,
    productName VARCHAR(255),
    productDescription VARCHAR(255),
    productPrice DECIMAL(10, 2),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
-- +goose Down
DROP TABLE IF EXISTS products;