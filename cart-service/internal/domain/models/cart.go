package models

// type CartItem struct {
// 	ID       string  `json:"id"`
// 	Name     string  `json:"name"`
// 	Quantity int     `json:"quantity"`
// 	Price    float64 `json:"price"`
// }

type CartItem struct {
	ID        string `json:"id"`
	UserID    int64  `json:"user_id"`
	ProductID int64  `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type Cart struct {
	Items []CartItem `json:"items"`
}
