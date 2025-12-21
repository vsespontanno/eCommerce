package saga

import (
	"context"
	"fmt"
	"sort"

	"github.com/vsespontanno/eCommerce/services/saga-orchestrator/internal/config"
	orderEntity "github.com/vsespontanno/eCommerce/services/saga-orchestrator/internal/domain/event/entity"
	"github.com/vsespontanno/eCommerce/services/saga-orchestrator/internal/domain/product/entity"
	"go.uber.org/zap"
)

type Step int

const (
	StepWallet Step = iota + 1
	StepProducts
)

type Eventer interface {
	ProccessEvent(ctx context.Context, event orderEntity.OrderEvent) error
}

type MoneyReserver interface {
	ReserveFunds(ctx context.Context, userID int64, amount int64) (string, error)
	CommitFunds(ctx context.Context, userID int64, amount int64) (string, error)
	ReleaseFunds(ctx context.Context, userID int64, amount int64) (string, error)
}

type ProductsReserver interface {
	ReserveProducts(ctx context.Context, productIDs []entity.Product) (bool, error)
	CommitProducts(ctx context.Context, productIDs []entity.Product) (bool, error)
	ReleaseProducts(ctx context.Context, productIDs []entity.Product) (bool, error)
}

type OutboxRepo interface {
	SaveEvent(ctx context.Context, event orderEntity.OrderEvent) error
}

type Orchestrator struct {
	config   *config.Config
	logger   *zap.SugaredLogger
	wallet   MoneyReserver
	products ProductsReserver
	outboxer OutboxRepo
}

func New(config *config.Config, wallet MoneyReserver, products ProductsReserver, outboxer OutboxRepo, logger *zap.SugaredLogger) *Orchestrator {
	return &Orchestrator{config: config, logger: logger, wallet: wallet, products: products, outboxer: outboxer}
}

func (o *Orchestrator) SagaTransaction(ctx context.Context, Order orderEntity.OrderEvent) error {
	// Шаг 1: Резервируем деньги
	_, err := o.wallet.ReserveFunds(ctx, Order.UserID, Order.Total)
	if err != nil {
		o.logger.Errorw("Failed to reserve funds", "error", err, "userID", Order.UserID, "amount", Order.Total)
		o.rollbackTransaction(ctx, Order, StepWallet)
		return fmt.Errorf("wallet reserve failed: %w", err)
	}

	// Сортируем товары по ID для предотвращения deadlock
	sort.Slice(Order.Products, func(i, j int) bool {
		return Order.Products[i].ID < Order.Products[j].ID
	})

	// Шаг 2: Резервируем товары
	_, err = o.products.ReserveProducts(ctx, Order.Products)
	if err != nil {
		o.logger.Errorw("Failed to reserve products", "error", err, "orderID", Order.OrderID)
		o.rollbackTransaction(ctx, Order, StepProducts)
		return fmt.Errorf("products reserve failed: %w", err)
	}

	// Шаг 3: Коммитим деньги
	_, err = o.wallet.CommitFunds(ctx, Order.UserID, Order.Total)
	if err != nil {
		o.logger.Errorw("Failed to commit funds", "error", err, "orderID", Order.OrderID)
		o.rollbackTransaction(ctx, Order, StepProducts)
		return fmt.Errorf("wallet commit failed: %w", err)
	}

	// Шаг 4: Коммитим товары
	_, err = o.products.CommitProducts(ctx, Order.Products)
	if err != nil {
		o.logger.Errorw("Failed to commit products", "error", err, "orderID", Order.OrderID)
		// КРИТИЧНО: Если commit товаров упал, нужно откатить commit денег!
		// Но это уже сложная ситуация - деньги уже списаны
		o.logger.Errorw("CRITICAL: Funds committed but products commit failed - manual intervention required", "orderID", Order.OrderID)
		o.rollbackTransaction(ctx, Order, StepProducts)
		return fmt.Errorf("products commit failed: %w", err)
	}

	// Шаг 5: ТОЛЬКО ПОСЛЕ успешного commit отправляем в Kafka
	Order.Status = "Completed"
	err = o.outboxer.SaveEvent(ctx, Order)
	if err != nil {
		o.logger.Errorw("Failed to save event", "error", err, "orderID", Order.OrderID)
		o.rollbackTransaction(ctx, Order, StepProducts)
		return fmt.Errorf("failed to save event: %w", err)
	}

	o.logger.Infow("Saga transaction completed successfully", "orderID", Order.OrderID, "userID", Order.UserID)
	return nil
}

func (o *Orchestrator) rollbackTransaction(ctx context.Context, order orderEntity.OrderEvent, step Step) {
	o.logger.Infow("Starting rollback", "orderID", order.OrderID, "step", step)

	switch step {
	case StepWallet:
		// Отменяем только резерв денег
		if _, err := o.wallet.ReleaseFunds(ctx, order.UserID, order.Total); err != nil {
			o.logger.Errorw("rollback: failed to release funds", "orderID", order.OrderID, "error", err)
		} else {
			o.logger.Infow("rollback: funds released successfully", "orderID", order.OrderID)
		}

	case StepProducts:
		// Отменяем резерв товаров
		if _, err := o.products.ReleaseProducts(ctx, order.Products); err != nil {
			o.logger.Errorw("rollback: failed to release products", "orderID", order.OrderID, "error", err)
		} else {
			o.logger.Infow("rollback: products released successfully", "orderID", order.OrderID)
		}

		// Отменяем резерв денег
		if _, err := o.wallet.ReleaseFunds(ctx, order.UserID, order.Total); err != nil {
			o.logger.Errorw("rollback: failed to release funds", "orderID", order.OrderID, "error", err)
		} else {
			o.logger.Infow("rollback: funds released successfully", "orderID", order.OrderID)
		}
	}

	o.logger.Infow("Rollback completed", "orderID", order.OrderID)
}
