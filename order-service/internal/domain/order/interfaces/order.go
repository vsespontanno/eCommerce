package interfaces

import (
	"context"

	"github.com/vsespontanno/eCommerce/order-service/internal/domain/order/entity"
)

type CartItem struct {
	ProductID string
	Quantity  int
}

// OrderRepository определяет контракт для хранения заказов
type OrderRepository interface {
	Save(ctx context.Context, order *entity.Order) error
	UpdateStatus(ctx context.Context, orderID string, status entity.OrderStatus) error
	Get(ctx context.Context, orderID string) (*entity.Order, error)
	Close() error
}
