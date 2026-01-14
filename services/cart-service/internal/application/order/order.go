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

type Client interface {
	CreateOrder(ctx context.Context, order *entity.OrderEvent) (string, error)
}

type CompleteService struct {
	logger       *zap.SugaredLogger
	pgCleaner    PGCartCleaner
	redisCleaner RedisCartCleaner
	orderClient  Client
}

func NewOrderCompleteService(logger *zap.SugaredLogger, pgCleaner PGCartCleaner, redisCleaner RedisCartCleaner, orderClient Client) *CompleteService {
	return &CompleteService{
		logger:       logger,
		pgCleaner:    pgCleaner,
		redisCleaner: redisCleaner,
		orderClient:  orderClient,
	}
}

func (o *CompleteService) CompleteOrder(ctx context.Context, order *entity.OrderEvent) error {
	o.logger.Infow("Processing order completion",
		"orderID", order.OrderID,
		"userID", order.UserID,
		"status", order.Status,
		"eventType", order.EventType,
	)

	// Шаг 1: Создаем заказ в order-service
	// ВАЖНО: order-service должен быть идемпотентным и проверять существование заказа по OrderID
	orderID, err := o.orderClient.CreateOrder(ctx, order)
	if err != nil {
		o.logger.Errorw("Failed to create order in order-service",
			"orderID", order.OrderID,
			"userID", order.UserID,
			"error", err,
		)
		return err
	}
	o.logger.Infow("Order created in order-service",
		"orderID", orderID,
		"userID", order.UserID,
	)

	// Шаг 2: Очищаем корзину в Postgres
	err = o.pgCleaner.CleanCart(ctx, order)
	if err != nil {
		o.logger.Errorw("Failed to clean cart in Postgres",
			"orderID", order.OrderID,
			"userID", order.UserID,
			"error", err,
		)
		// Заказ уже создан, но корзина не очищена
		// Возвращаем ошибку чтобы Kafka consumer не закоммитил offset
		// и сообщение обработалось повторно
		return err
	}

	// Шаг 3: Очищаем корзину в Redis
	err = o.redisCleaner.CleanCart(ctx, order)
	if err != nil {
		o.logger.Errorw("Failed to clean redis cart",
			"orderID", order.OrderID,
			"userID", order.UserID,
			"error", err,
		)
		// Не критично если Redis не очистился - корзина синхронизируется из Postgres
		// Не возвращаем ошибку, чтобы не блокировать обработку
	}

	o.logger.Infow("Order completed successfully",
		"orderID", order.OrderID,
		"userID", order.UserID,
		"total", order.Total,
	)
	return nil
}
