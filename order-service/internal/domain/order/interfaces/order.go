package interfaces

import (
	"context"

	"github.com/vsespontanno/eCommerce/order-service/internal/domain/order/entity"
)

type OrderRepo interface {
	CreateOrder(ctx context.Context, order *entity.Order) (string, error)
	GetOrder(ctx context.Context, orderID string) (*entity.Order, error)
	ListOrdersByUser(ctx context.Context, userID int64, limit, offset int) ([]entity.Order, error)
}
