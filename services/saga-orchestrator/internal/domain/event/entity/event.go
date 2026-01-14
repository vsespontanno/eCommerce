package entity

import "github.com/vsespontanno/eCommerce/services/saga-orchestrator/internal/domain/product/entity"

// EventType константы для типов событий
const (
	EventTypeOrderCompleted = "OrderCompleted"
	EventTypeOrderFailed    = "OrderFailed"
	EventTypeOrderCancelled = "OrderCancelled"
)

type OrderEvent struct {
	OrderID   string           `json:"order_id"`
	UserID    int64            `json:"user_id"`
	Products  []entity.Product `json:"products"`
	Total     int64            `json:"total"`
	Status    string           `json:"status"`
	EventType string           `json:"event_type,omitempty"` // Тип события для routing в consumer
}
