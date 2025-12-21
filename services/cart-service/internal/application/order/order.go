package order

import (
	"context"

	"github.com/vsespontanno/eCommerce/services/cart-service/internal/domain/order/entity"
	"go.uber.org/zap"
)

type PGCartCleaner interface {
	CleanCart(ctx context.Context, order *entity.OrderEvent) error
}

type RedisCartCleaner interface {
	CleanCart(ctx context.Context, order *entity.OrderEvent) error
}

type OrderClient interface {
	CreateOrder(ctx context.Context, order *entity.OrderEvent) (string, error)
}

type OrderCompleteService struct {
	logger       *zap.SugaredLogger
	pgCleaner    PGCartCleaner
	redisCleaner RedisCartCleaner
	orderClient  OrderClient
}

func NewOrderCompleteService(logger *zap.SugaredLogger, pgCleaner PGCartCleaner, redisCleaner RedisCartCleaner, orderClient OrderClient) *OrderCompleteService {
	return &OrderCompleteService{
		logger:       logger,
		pgCleaner:    pgCleaner,
		redisCleaner: redisCleaner,
		orderClient:  orderClient,
	}
}

func (o *OrderCompleteService) CompleteOrder(ctx context.Context, order *entity.OrderEvent) error {
	// Шаг 1: Создаем заказ в order-service
	orderID, err := o.orderClient.CreateOrder(ctx, order)
	if err != nil {
		o.logger.Errorw("Failed to create order in order-service", "orderID", order.OrderID, "error", err)
		return err
	}
	o.logger.Infow("Order created in order-service", "orderID", orderID)

	// Шаг 2: Очищаем корзину в Postgres
	err = o.pgCleaner.CleanCart(ctx, order)
	if err != nil {
		o.logger.Errorw("Failed to clean cart in Postgres", "orderID", order.OrderID, "error", err)
		// Заказ уже создан, но корзина не очищена - логируем как warning
		// В идеале нужен механизм компенсации или retry
		return err
	}

	// Шаг 3: Очищаем корзину в Redis
	err = o.redisCleaner.CleanCart(ctx, order)
	if err != nil {
		o.logger.Errorw("Failed to clean redis cart", "orderID", order.OrderID, "error", err)
		// Не критично если Redis не очистился - корзина синхронизируется из Postgres
		// Но логируем ошибку
	}

	o.logger.Infow("Order completed successfully", "orderID", order.OrderID, "userID", order.UserID)
	return nil
}
