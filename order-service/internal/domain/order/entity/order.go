package entity

import "time"

type OrderItem struct {
	ProductID int64 `db:"product_id" json:"product_id"`
	Price     int64 `db:"price" json:"price"`
	Quantity  int64 `db:"quantity" json:"quantity"`
}

type Order struct {
	ID        string      `db:"id" json:"order_id"`
	UserID    int64       `db:"user_id" json:"user_id"`
	Total     int64       `db:"total" json:"total"`
	Status    string      `db:"status" json:"status"`
	Items     []OrderItem `json:"items"`
	CreatedAt time.Time   `db:"created_at" json:"created_at"`
}
