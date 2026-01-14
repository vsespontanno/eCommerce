package order

import (
	"context"

	"github.com/vsespontanno/eCommerce/services/order-service/internal/domain/order/entity"
	"github.com/vsespontanno/eCommerce/services/order-service/internal/domain/order/interfaces"
	"go.uber.org/zap"
)

type Service struct {
	repo   interfaces.OrderRepo
	logger *zap.SugaredLogger
}

func NewOrderService(repo interfaces.OrderRepo, logger *zap.SugaredLogger) *Service {
	return &Service{repo: repo, logger: logger}
}

// Called by Cart Service when order is confirmed by Saga
func (s *Service) CreateOrder(ctx context.Context, order *entity.Order) (string, error) {
	s.logger.Infow("Creating order",
		"order_id", order.OrderID,
		"user_id", order.UserID,
		"total", order.Total,
		"status", order.Status,
		"items_count", len(order.Products),
	)

	err := s.repo.CreateOrder(ctx, order)
	if err != nil {
		s.logger.Errorw("Failed to create order in repository",
			"order_id", order.OrderID,
			"user_id", order.UserID,
			"error", err,
		)
		return "", err
	}

	s.logger.Infow("Order created successfully",
		"order_id", order.OrderID,
		"user_id", order.UserID,
	)
	return order.OrderID, nil
}

func (s *Service) GetOrder(ctx context.Context, orderID string) (*entity.Order, error) {
	return s.repo.GetOrder(ctx, orderID)
}

func (s *Service) ListOrdersByUser(ctx context.Context, userID int64, limit, offset uint64) ([]entity.Order, error) {
	return s.repo.ListOrdersByUser(ctx, userID, limit, offset)
}
