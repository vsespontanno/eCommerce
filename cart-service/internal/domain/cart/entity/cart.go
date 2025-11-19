package entity

type CartItem struct {
	UserID    int64 `json:"user_id"`
	ProductID int64 `json:"product_id"`
	Quantity  int64 `json:"quantity"`
	Price     int64 `json:"price"`
}

type Cart struct {
	Items []CartItem `json:"items"`
}
