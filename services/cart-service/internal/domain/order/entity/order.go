package entity

type OrderEvent struct {
	OrderID  string            `json:"order_id"`
	UserID   int64             `json:"user_id"`
	Products []ProductForOrder `json:"products"`
	Total    int64             `json:"total"`
	Status   string            `json:"status"`
}

type ProductForOrder struct {
	ID       int64 `json:"product_id"`
	Quantity int64 `json:"quantity"`
}
