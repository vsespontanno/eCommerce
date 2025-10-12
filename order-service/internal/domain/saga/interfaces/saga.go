package interfaces

import (
	"context"

	"github.com/vsespontanno/eCommerce/order-service/internal/domain/saga/entity"
)

type SagaRepo interface {
	Create(ctx context.Context, s *entity.Saga) error
	Update(ctx context.Context, s *entity.Saga) error
	Get(ctx context.Context, id string) (*entity.Saga, error)
}
