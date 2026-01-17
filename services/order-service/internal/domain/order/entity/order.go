package entity

import "time"

type OrderItem struct {
	ProductID int64 `db:"product_id" json:"product_id"`
	Quantity  int64 `db:"quantity" json:"quantity"`
}

type Order struct {
	OrderID   string      `db:"order_id" json:"order_id"`
	UserID    int64       `db:"user_id" json:"user_id"`
	Products  []OrderItem `json:"products"`
	Total     int64       `db:"total" json:"total"`
	Status    string      `db:"status" json:"status"`
	CreatedAt time.Time   `db:"created_at" json:"created_at,omitempty"`
}
