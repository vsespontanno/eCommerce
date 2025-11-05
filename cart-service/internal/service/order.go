package service

import (
	"context"

	"github.com/vsespontanno/eCommerce/cart-service/internal/domain/models"
	"github.com/vsespontanno/eCommerce/cart-service/internal/repository/redis"
	"go.uber.org/zap"
)

type Producter interface {
	GetProduct(ctx context.Context, productID int64) (*models.Product, error)
}

type OrderService struct {
	sugarLogger   *zap.SugaredLogger
	redisStore    *redis.OrderStore
	productClient Producter
}

func NewOrder(logger *zap.SugaredLogger, redisStore *redis.OrderStore, productClient Producter) *OrderService {
	return &OrderService{
		sugarLogger:   logger,
		redisStore:    redisStore,
		productClient: productClient,
	}
}

func (s *OrderService) AddProductToCart(ctx context.Context, userID int64, productID int64) error {
	q, err := s.redisStore.GetProduct(ctx, userID, productID)
	if err != nil {
		if err != models.ErrProductIsNotInCart {
			s.sugarLogger.Errorf("error while getting and adding 1 product to cart: %w", err)
			return err
		}
		product, err := s.productClient.GetProduct(ctx, productID)
		if err != nil {
			s.sugarLogger.Errorf("error while getting product from grpc-client and adding 1 product to cart: %w", err)
			return err
		}
		return s.redisStore.AddNewProductToCart(ctx, userID, product)
	}
	if q.Quantity == 100 {
		return models.ErrTooManyProductsOfOneType
	}
	err = s.redisStore.IncrementInCart(ctx, userID, productID)
	if err != nil {
		s.sugarLogger.Errorf("error while incrementing 1 product to cart: %w", err)
	}
	return err
}

func (s *OrderService) DeleteProductFromCart(ctx context.Context, userID int64, productID int64) error {
	err := s.redisStore.DecrementInCart(ctx, userID, productID)
	if err != nil {
		s.sugarLogger.Errorf("error while deleting 1 product to cart: %w", err)
	}
	return err
}
