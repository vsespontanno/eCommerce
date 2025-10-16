package models

type CartItem struct {
	ID        string `json:"id"`
	UserID    int64  `json:"user_id"`
	ProductID int64  `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type Cart struct {
	Items []CartItem `json:"items"`
}

type SagaCart struct {
	UserID int64 `json:"user_id"`
	Order  Items `json:"order"`
}

type Items struct {
	ProductID int64 `json:"product_id"`
	Quantity  int   `json:"quantity"`
}
