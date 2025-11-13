package models

type CartItem struct {
	ID        int64 `json:"id"`
	UserID    int64 `json:"user_id"`
	ProductID int64 `json:"product_id"`
	Quantity  int   `json:"quantity"`
}

type Cart struct {
	Items []CartItem `json:"items"`
}
