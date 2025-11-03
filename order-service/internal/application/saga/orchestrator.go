package saga

import (
	"context"
	"fmt"
	"sort"

	"github.com/vsespontanno/eCommerce/order-service/internal/config"
	orderEntity "github.com/vsespontanno/eCommerce/order-service/internal/domain/event/entity"
)

type MoneyReserver interface {
	ReserveFunds(ctx context.Context, userID int64, amount int64) (string, error)
	CommitFunds(ctx context.Context, userID int64, amount int64) (string, error)
	ReleaseFunds(ctx context.Context, userID int64, amount int64) (string, error)
}

type Orchestrator struct {
	config *config.Config
	wallet MoneyReserver
}

func New(config *config.Config) *Orchestrator {
	return &Orchestrator{config: config}
}

// TODO: product reserve; release; generic func to release if something's wrong;
func (o *Orchestrator) SagaTransaction(ctx context.Context, Order orderEntity.OrderEvent) error {
	response, err := o.wallet.ReserveFunds(ctx, Order.UserID, Order.Total)
	if err != nil {
		if response != "" {
			return fmt.Errorf("Not enough money to make an order")
		}
		return err
	}
	sort.Slice(Order.Products, func(i, j int) bool {
		return Order.Products[i].ID < Order.Products[j].ID
	})
	return nil
}
