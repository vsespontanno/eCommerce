package interfaces

import (
	"context"

	"github.com/vsespontanno/eCommerce/services/order-service/internal/domain/order/entity"
)

type OrderRepo interface {
	CreateOrder(ctx context.Context, order *entity.Order) error
	GetOrder(ctx context.Context, orderID string) (*entity.Order, error)
	ListOrdersByUser(ctx context.Context, userID int64, limit, offset uint64) ([]entity.Order, error)
}
