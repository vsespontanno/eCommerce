package service

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/vsespontanno/eCommerce/cart-service/internal/domain/models"
	"go.uber.org/zap"
)

type Producter interface {
	Product(ctx context.Context, productID int64) (*models.Product, error)
}

type RedisCartRepo interface {
	AddNewProductToCart(ctx context.Context, userID int64, product *models.Product) error
	SaveCart(ctx context.Context, userID int64, cart *models.Cart) error
	DecrementInCart(ctx context.Context, userID int64, productID int64) error
	GetCart(ctx context.Context, userID int64) (*models.Cart, error)
	GetProduct(ctx context.Context, userID int64, productID int64) (*models.Product, error)
	IncrementInCart(ctx context.Context, userID int64, productID int64) error
	RemoveProductFromCart(ctx context.Context, userID int64, productID int64) error
	DeleteProduct(ctx context.Context, userID int64, productID int64) error
	ClearCart(ctx context.Context, userID int64) error
}

type PostgresCartRepo interface {
	GetCart(ctx context.Context, userID int64) (*models.Cart, error)
}

type CartService struct {
	sugarLogger   *zap.SugaredLogger
	redisStore    RedisCartRepo
	productClient Producter
	cartStore     PostgresCartRepo
}

func NewCart(logger *zap.SugaredLogger, redisStore RedisCartRepo, productClient Producter, cartStore PostgresCartRepo) *CartService {
	return &CartService{
		sugarLogger:   logger,
		redisStore:    redisStore,
		productClient: productClient,
		cartStore:     cartStore,
	}
}

func (s *CartService) Cart(ctx context.Context, userID int64) (*models.Cart, error) {
	cart, err := s.redisStore.GetCart(ctx, userID)
	if err != nil {
		if err == redis.Nil {
			dbCart, err := s.cartStore.GetCart(ctx, userID)
			if err != nil {
				if err == models.ErrNoCartFound {
					return &models.Cart{}, nil
				}
				s.sugarLogger.Errorf("error while getting cart from postgres: %w", err)
				return nil, err
			}
			err = s.redisStore.SaveCart(ctx, userID, dbCart)
			if err != nil {
				s.sugarLogger.Errorf("error while saving cart to redis: %w", err)
				return nil, err
			}
			return dbCart, nil
		}
		s.sugarLogger.Errorf("error while getting cart from store: %w", err)
		return nil, err
	}
	return cart, nil
}

func (s *CartService) ClearCart(ctx context.Context, userID int64) error {
	err := s.redisStore.ClearCart(ctx, userID)
	if err != nil {
		s.sugarLogger.Errorf("error while clearing cart: %w", err)
	}
	return err
}
func (s *CartService) AddProductToCart(ctx context.Context, userID int64, productID int64) error {
	q, err := s.redisStore.GetProduct(ctx, userID, productID)
	if err != nil {
		if err != models.ErrProductIsNotInCart {
			s.sugarLogger.Errorf("error while getting and adding 1 product to cart: %w", err)
			return err
		}
		product, err := s.productClient.Product(ctx, productID)
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

func (s *CartService) Increment(ctx context.Context, userID int64, productID int64) error {
	err := s.redisStore.IncrementInCart(ctx, userID, productID)
	if err != nil {
		s.sugarLogger.Errorf("error while incrementing 1 product to cart: %w", err)
	}
	return err
}

func (s *CartService) Decrement(ctx context.Context, userID int64, productID int64) error {
	err := s.redisStore.DecrementInCart(ctx, userID, productID)
	if err != nil {
		s.sugarLogger.Errorf("error while decrementing 1 product to cart: %w", err)
	}
	return err
}

func (s *CartService) DeleteProductFromCart(ctx context.Context, userID int64, productID int64) error {
	err := s.redisStore.DeleteProduct(ctx, userID, productID)
	if err != nil {
		s.sugarLogger.Errorf("error while deleting 1 product to cart: %w", err)
	}
	return err
}
