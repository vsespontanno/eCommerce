package service

import (
	"context"
	"fmt"

	"github.com/vsespontanno/eCommerce/cart-service/internal/domain/models"
	"go.uber.org/zap"
)

type Saga interface {
	StartCheckout(ctx context.Context, userID int64, cart []models.Product) (bool, error)
}

type Carter interface {
	GetCart(ctx context.Context, userID int64) ([]models.Product, error)
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

func (s *SagaService) Checkout(ctx context.Context, userID int64) (bool, error) {
	cart, err := s.redisStore.GetCart(ctx, userID)
	if err != nil {
		s.sugarLogger.Errorf("error while getting cart from store: %v", err)
		return false, err
	}
	fmt.Println(cart)
	ok, err := s.sagaClient.StartCheckout(ctx, userID, cart)
	if err != nil {
		s.sugarLogger.Errorf("error while starting checkout: %v", err)
		return false, err
	}

	return ok, nil

}
