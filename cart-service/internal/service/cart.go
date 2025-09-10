package service

import (
	"context"

	"github.com/vsespontanno/eCommerce/cart-service/internal/domain/models"
	"github.com/vsespontanno/eCommerce/cart-service/internal/repository/postgres"
	"go.uber.org/zap"
)

type CartService struct {
	sugarLogger *zap.SugaredLogger
	cartStore   *postgres.CartStore
}

func NewCart(sugarLogger *zap.SugaredLogger, cartStore *postgres.CartStore) *CartService {
	return &CartService{
		sugarLogger: sugarLogger,
		cartStore:   cartStore,
	}
}

func (s *CartService) Cart(ctx context.Context, userID int64) (*models.Cart, error) {
	cart, err := s.cartStore.GetCart(ctx, userID)
	if err != nil {
		s.sugarLogger.Errorf("error while getting cart from store: %w", err)
		return &models.Cart{}, err
	}
	return cart, err
}


