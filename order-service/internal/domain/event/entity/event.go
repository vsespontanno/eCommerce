package entity

import "github.com/vsespontanno/eCommerce/order-service/internal/domain/product/entity"

type OrderEvent struct {
	OrderID  string           `json:"order_id"`
	UserID   string           `json:"user_id"`
	Products []entity.Product `json:"products"`
	Total    int              `json:"total"`
	Status   string           `json:"status"`
}
