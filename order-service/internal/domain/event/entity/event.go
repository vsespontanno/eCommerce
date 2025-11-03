package entity

import "github.com/vsespontanno/eCommerce/order-service/internal/domain/product/entity"

type OrderEvent struct {
	OrderID  string
	UserID   int64
	Products []entity.Product
	Total    int64
	Status   string
}
