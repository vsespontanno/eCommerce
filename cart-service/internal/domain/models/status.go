package models

type OrderStatus string

const (
	StatusDraft     OrderStatus = "draft"     // Черновик
	StatusReserved  OrderStatus = "reserved"  // Товары зарезервированы
	StatusPaid      OrderStatus = "paid"      // Оплачено
	StatusShipped   OrderStatus = "shipped"   // Отправлено
	StatusDelivered OrderStatus = "delivered" // Доставлено
	StatusCancelled OrderStatus = "cancelled" // Отменено
)

type SelectedProduct struct {
	ProductID int64 `json:"product_id"`
	Quantity  int64 `json:"quantity"`
}

type SelectedCart struct {
	UserID int64             `json:"user_id"`
	Items  []SelectedProduct `json:"items"`
	Status OrderStatus       `json:"status"`
}

type CheckoutRequest struct {
	UserID int64 `json:"user_id"`
}

type CheckoutResponse struct {
	OrderID string      `json:"order_id"`
	Status  OrderStatus `json:"status"`
	Message string      `json:"message"`
}
