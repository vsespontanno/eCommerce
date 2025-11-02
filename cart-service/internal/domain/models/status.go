package models

type OrderEvent struct {
	OrderID  string
	UserID   string
	Products []Product
	Total    int
	Status   string
}

type Product struct {
	ID       string
	Quantity int
}
