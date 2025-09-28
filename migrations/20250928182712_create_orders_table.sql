-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY, 
    user_id UUID, 
    total INT, 
    status TEXT
); 

CREATE TABLE order_items (
    order_id UUID, 
    product_id UUID, 
    quantity INT NOT NULL DEFAULT 15, 
    price INT NOT NULL DEFAULT 12
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
