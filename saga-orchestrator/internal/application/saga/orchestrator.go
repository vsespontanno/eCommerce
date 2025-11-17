package saga

import (
	"context"
	"fmt"
	"sort"

	"github.com/vsespontanno/eCommerce/saga-orchestrator/internal/config"
	orderEntity "github.com/vsespontanno/eCommerce/saga-orchestrator/internal/domain/event/entity"
	"github.com/vsespontanno/eCommerce/saga-orchestrator/internal/domain/product/entity"
	"go.uber.org/zap"
)

type SagaStep int

const (
	StepWallet SagaStep = iota + 1
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

type Orchestrator struct {
	config   *config.Config
	logger   *zap.SugaredLogger
	wallet   MoneyReserver
	products ProductsReserver
	eventer  Eventer
}

func New(config *config.Config, wallet MoneyReserver, products ProductsReserver, eventer Eventer, logger *zap.SugaredLogger) *Orchestrator {
	return &Orchestrator{config: config, logger: logger, wallet: wallet, products: products, eventer: eventer}
}

// TODO: product reserve; release; generic func to release if something's wrong;
func (o *Orchestrator) SagaTransaction(ctx context.Context, Order orderEntity.OrderEvent) error {
	_, err := o.wallet.ReserveFunds(ctx, Order.UserID, Order.Total)
	if err != nil {
		o.rollbackTransaction(ctx, Order, StepWallet)
		return fmt.Errorf("wallet reserve failed: %w", err)
	}

	sort.Slice(Order.Products, func(i, j int) bool {
		return Order.Products[i].ID < Order.Products[j].ID
	})

	_, err = o.products.ReserveProducts(ctx, Order.Products)
	if err != nil {
		o.logger.Errorw("Error reserving products", "error", err, "stage", "Orchestrator.SagaTransaction", "step", 1)
		o.rollbackTransaction(ctx, Order, StepProducts)
		return err
	}

	err = o.eventer.ProccessEvent(ctx, Order)
	if err != nil {
		o.logger.Errorw("Failed to publish Kafka message", "orderID", Order.OrderID, "error", err)
		o.rollbackTransaction(ctx, Order, StepProducts)
		return err
	}

	_, err = o.wallet.CommitFunds(ctx, Order.UserID, Order.Total)
	if err != nil {
		o.logger.Errorw("Error committing funds", "error", err, "stage", "Orchestrator.SagaTransaction", "step", 2)
		o.rollbackTransaction(ctx, Order, StepProducts)
		return err
	}
	_, err = o.products.CommitProducts(ctx, Order.Products)
	if err != nil {
		o.logger.Errorw("Error committing products", "error", err, "stage", "Orchestrator.SagaTransaction", "step", 3)
		o.rollbackTransaction(ctx, Order, StepProducts)
		return err
	}
	return nil
}

func (o *Orchestrator) rollbackTransaction(ctx context.Context, Order orderEntity.OrderEvent, step SagaStep) {
	switch step {
	case 1:
		// Отменяем только резерв бабок
		if _, err := o.wallet.ReleaseFunds(ctx, Order.UserID, Order.Total); err != nil {
			o.logger.Errorw("rollback: failed to release funds", "error", err)
		}
	case 2:
		// Отменяем всё
		if _, err := o.products.ReleaseProducts(ctx, Order.Products); err != nil {
			o.logger.Errorw("rollback: failed to release products", "error", err)
		}
		if _, err := o.wallet.ReleaseFunds(ctx, Order.UserID, Order.Total); err != nil {
			o.logger.Errorw("rollback: failed to release funds", "error", err)
		}
	}
}
