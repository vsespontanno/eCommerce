package order

import (
	"context"

	"github.com/vsespontanno/eCommerce/cart-service/internal/domain/order/entity"
	"go.uber.org/zap"
)

type PGCartCleaner interface {
	CleanCart(ctx context.Context, order *entity.OrderEvent) error
}

type RedisCartCleaner interface {
	CleanCart(ctx context.Context, order *entity.OrderEvent) error
}

type OrderCompleteService struct {
	logger       *zap.SugaredLogger
	pgCleaner    PGCartCleaner
	redisCleaner RedisCartCleaner
}

func NewOrderCompleteService(logger *zap.SugaredLogger, pgCleaner PGCartCleaner, redisCleaner RedisCartCleaner) *OrderCompleteService {
	return &OrderCompleteService{
		logger:       logger,
		pgCleaner:    pgCleaner,
		redisCleaner: redisCleaner,
	}
}

func (o *OrderCompleteService) CompleteOrder(ctx context.Context, order *entity.OrderEvent) error {
	order.Status = "Completed"
	err := o.pgCleaner.CleanCart(ctx, order)
	if err != nil {
		o.logger.Errorw("Failed to clean cart", "error", err)
		return err
	}
	err = o.redisCleaner.CleanCart(ctx, order)
	if err != nil {
		o.logger.Errorw("Failed to clean redis cart", "error", err)
		return err
	}
	return nil
}
