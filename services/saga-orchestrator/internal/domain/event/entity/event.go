package entity

import "github.com/vsespontanno/eCommerce/services/saga-orchestrator/internal/domain/product/entity"

type OrderEvent struct {
	OrderID  string           `json:"order_id"`
	UserID   int64            `json:"user_id"`
	Products []entity.Product `json:"products"`
	Total    int64            `json:"total"`
	Status   string           `json:"status"`
}
