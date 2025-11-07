package models

type OrderEvent struct {
	OrderID  string            `json:"order_id"`
	UserID   string            `json:"user_id"`
	Products []ProductForOrder `json:"products"`
	Total    int               `json:"total"`
	Status   string            `json:"status"`
}

type ProductForOrder struct {
	ID       string `json:"product_id"`
	Quantity int    `json:"quantity"`
}
