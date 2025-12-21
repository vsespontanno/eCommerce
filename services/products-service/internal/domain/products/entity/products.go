package entity

type Product struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Price        int64  `json:"price"`
	Description  string `json:"description"`
	Category     string `json:"category"`
	Brand        string `json:"brand"`
	Rating       int    `json:"rating"`
	NumReviews   int    `json:"num_reviews"`
	CreatedAt    string `json:"created_at"`
	CountInStock int    `json:"count_in_stock"`
}
