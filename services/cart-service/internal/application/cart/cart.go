package cart

import (
	"context"

	"github.com/vsespontanno/eCommerce/services/cart-service/internal/domain/apperrors"
	"github.com/vsespontanno/eCommerce/services/cart-service/internal/domain/cart/entity"
	"go.uber.org/zap"
)

type Producter interface {
	Product(ctx context.Context, productID int64) (*entity.CartItem, error)
}

type RedisCartRepo interface {
	AddNewProductToCart(ctx context.Context, userID int64, product *entity.CartItem) error
	SaveCart(ctx context.Context, userID int64, cart *entity.Cart) error
	DecrementInCart(ctx context.Context, userID int64, productID int64) error
	GetCart(ctx context.Context, userID int64) (*entity.Cart, error)
	GetProduct(ctx context.Context, userID int64, productID int64) (*entity.CartItem, error)
	IncrementInCart(ctx context.Context, userID int64, productID int64) error
	RemoveProductFromCart(ctx context.Context, userID int64, productID int64) error
	DeleteProduct(ctx context.Context, userID int64, productID int64) error
	ClearCart(ctx context.Context, userID int64) error
}

type PostgresCartRepo interface {
	GetCart(ctx context.Context, userID int64) (*entity.Cart, error)
}

type Service struct {
	sugarLogger        *zap.SugaredLogger
	redisStore         RedisCartRepo
	productClient      Producter
	cartStore          PostgresCartRepo
	maxProductQuantity int
}

func NewCart(logger *zap.SugaredLogger, redisStore RedisCartRepo, productClient Producter, cartStore PostgresCartRepo, maxProductQuantity int) *Service {
	return &Service{
		sugarLogger:        logger,
		redisStore:         redisStore,
		productClient:      productClient,
		cartStore:          cartStore,
		maxProductQuantity: maxProductQuantity,
	}
}

func (s *Service) Cart(ctx context.Context, userID int64) (*entity.Cart, error) {
	cart, err := s.redisStore.GetCart(ctx, userID)
	if err != nil {
		if err == apperrors.ErrNoCartFound {
			dbCart, getErr := s.cartStore.GetCart(ctx, userID)
			if getErr != nil {
				if getErr == apperrors.ErrNoCartFound {
					return &entity.Cart{}, getErr
				}
				s.sugarLogger.Errorf("error while getting cart from postgres: %w", getErr)
				return nil, getErr
			}
			err = s.redisStore.SaveCart(ctx, userID, dbCart)
			if err != nil {
				s.sugarLogger.Errorf("error while saving cart to redis: %w", err)
				return nil, err
			}
			return dbCart, nil
		}
		s.sugarLogger.Errorw("error while getting cart from store: %w", err)
		return nil, err
	}
	return cart, nil
}

func (s *Service) ClearCart(ctx context.Context, userID int64) error {
	err := s.redisStore.ClearCart(ctx, userID)
	if err != nil {
		s.sugarLogger.Errorf("error while clearing cart: %w", err)
	}
	return err
}
func (s *Service) AddProductToCart(ctx context.Context, userID int64, productID int64) error {
	q, err := s.redisStore.GetProduct(ctx, userID, productID)
	if err != nil {
		if err != apperrors.ErrProductIsNotInCart {
			s.sugarLogger.Errorf("error while getting and adding 1 product to cart: %w", err)
			return err
		}
		product, prodErr := s.productClient.Product(ctx, productID)
		if prodErr != nil {
			s.sugarLogger.Errorf("error while getting product from grpc-client and adding 1 product to cart: %w", prodErr)
			return prodErr
		}
		product.UserID = userID
		return s.redisStore.AddNewProductToCart(ctx, userID, product)
	}
	if q.Quantity >= int64(s.maxProductQuantity) {
		return apperrors.ErrTooManyProductsOfOneType
	}
	err = s.redisStore.IncrementInCart(ctx, userID, productID)
	if err != nil {
		s.sugarLogger.Errorf("error while incrementing 1 product to cart: %w", err)
	}
	return err
}

func (s *Service) Increment(ctx context.Context, userID int64, productID int64) error {
	err := s.redisStore.IncrementInCart(ctx, userID, productID)
	if err != nil {
		s.sugarLogger.Errorf("error while incrementing 1 product to cart: %w", err)
	}
	return err
}

func (s *Service) Decrement(ctx context.Context, userID int64, productID int64) error {
	err := s.redisStore.DecrementInCart(ctx, userID, productID)
	if err != nil {
		s.sugarLogger.Errorf("error while decrementing 1 product to cart: %w", err)
	}
	return err
}

func (s *Service) DeleteProductFromCart(ctx context.Context, userID int64, productID int64) error {
	err := s.redisStore.DeleteProduct(ctx, userID, productID)
	if err != nil {
		s.sugarLogger.Errorf("error while deleting 1 product to cart: %w", err)
	}
	return err
}
