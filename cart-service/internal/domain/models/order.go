package models

type OrderEvent struct {
	OrderID  string
	UserID   string
	Products []ProductForOrder
	Total    int
	Status   string
}

type ProductForOrder struct {
	ID       string
	Quantity int
}
