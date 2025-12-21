package saga

import (
	"context"

	"github.com/vsespontanno/eCommerce/services/cart-service/internal/domain/cart/entity"
	"go.uber.org/zap"
)

type Saga interface {
	StartCheckout(ctx context.Context, userID int64, cart *entity.Cart) (string, error)
}

type Carter interface {
	GetCartProducts(ctx context.Context, userID int64) (*entity.Cart, error)
}

type SagaService struct {
	sugarLogger *zap.SugaredLogger
	redisStore  Carter
	sagaClient  Saga
}

func NewSagaService(sugarLogger *zap.SugaredLogger, redisStore Carter, sagaClient Saga) *SagaService {
	return &SagaService{
		sugarLogger: sugarLogger,
		redisStore:  redisStore,
		sagaClient:  sagaClient,
	}
}

func (s *SagaService) Checkout(ctx context.Context, userID int64) (string, error) {
	cart, err := s.redisStore.GetCartProducts(ctx, userID)
	if err != nil {
		s.sugarLogger.Errorf("error while getting cart from store: %v", err)
		return "", err
	}
	resp, err := s.sagaClient.StartCheckout(ctx, userID, cart)
	if err != nil {
		s.sugarLogger.Errorf("error while starting checkout: %v", err)
		return resp, err
	}

	return resp, nil

}
