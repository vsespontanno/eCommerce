package entity

type OrderItem struct {
	ProductID int64 `db:"product_id" json:"product_id"`
	Quantity  int64 `db:"quantity" json:"quantity"`
}

type Order struct {
	OrderID  string      `json:"order_id"`
	UserID   int64       `json:"user_id"`
	Products []OrderItem `json:"products"`
	Total    int64       `json:"total"`
	Status   string      `json:"status"`
}
