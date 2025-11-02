package service

import (
	"context"

	"github.com/vsespontanno/eCommerce/cart-service/internal/domain/models"
	"github.com/vsespontanno/eCommerce/cart-service/internal/repository/redis"
	"go.uber.org/zap"
)

type OrderService struct {
	sugarLogger *zap.SugaredLogger
	redisStore  *redis.OrderStore
}

func NewOrder(logger *zap.SugaredLogger, redisStore *redis.OrderStore) *OrderService {
	return &OrderService{
		sugarLogger: logger,
		redisStore:  redisStore,
	}
}

func (s *OrderService) AddAllProducts(ctx context.Context, userID int64, productIDs []int64) error {
	err := s.redisStore.AddAllProductsToCart(ctx, userID, productIDs)
	if err != nil {
		s.sugarLogger.Errorf("error while adding products to cart: %w", err)
	}
	return err
}

func (s *OrderService) AddProductToCart(ctx context.Context, userID int64, productID int64) error {
	q, err := s.redisStore.GetProductQuantity(ctx, userID, productID)
	if err != nil {
		if err != models.ErrProductIsNotInCart {
			return err
		}
	}
	if q == 100 {
		return models.ErrTooManyProductsOfOneType
	}
	err = s.redisStore.AddToCart(ctx, userID, productID)
	if err != nil {
		s.sugarLogger.Errorf("error while adding 1 product to cart: %w", err)
	}
	return err
}

func (s *OrderService) DeleteProductFromCart(ctx context.Context, userID int64, productID int64) error {
	err := s.redisStore.RemoveOneFromCart(ctx, userID, productID)
	if err != nil {
		s.sugarLogger.Errorf("error while deleting 1 product to cart: %w", err)
	}
	return err
}

// SAGA методы для работы с выбранными товарами
func (s *OrderService) GetSelectedProducts(ctx context.Context, userID int64) (map[int64]int64, error) {
	selected, err := s.redisStore.GetCart(ctx, userID)
	if err != nil {
		s.sugarLogger.Errorf("error while getting selected products: %w", err)
		return nil, err
	}
	return selected, nil
}
