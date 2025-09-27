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

func (s *OrderService) SelectProduct(ctx context.Context, userID int64, productID int64) error {
	// Проверяем, что товар есть в wishlist (через cartService)
	// Это будет проверяться в handler

	// Добавляем в selection (Redis)
	err := s.redisStore.AddToCart(ctx, userID, productID)
	if err != nil {
		s.sugarLogger.Errorf("error while selecting product: %w", err)
		return err
	}
	return nil
}

func (s *OrderService) UnselectProduct(ctx context.Context, userID int64, productID int64) error {
	err := s.redisStore.RemoveProductFromCart(ctx, userID, productID)
	if err != nil {
		s.sugarLogger.Errorf("error while unselecting product: %w", err)
		return err
	}
	return nil
}

// SAGA методы для резервирования и компенсации
func (s *OrderService) ReserveProducts(ctx context.Context, userID int64) error {
	// Здесь будет логика резервирования товаров
	// Пока просто логируем
	s.sugarLogger.Infof("Reserving products for user %d", userID)
	return nil
}

func (s *OrderService) ReleaseProducts(ctx context.Context, userID int64) error {
	// Компенсация - освобождаем резерв
	s.sugarLogger.Infof("Releasing products for user %d (compensation)", userID)
	return nil
}

func (s *OrderService) ConfirmOrder(ctx context.Context, userID int64) error {
	// Подтверждение заказа
	s.sugarLogger.Infof("Confirming order for user %d", userID)
	return nil
}

func (s *OrderService) CancelOrder(ctx context.Context, userID int64) error {
	// Отмена заказа - очищаем selection
	// Получаем все выбранные товары и удаляем их
	selected, err := s.redisStore.GetCart(ctx, userID)
	if err != nil {
		s.sugarLogger.Errorf("error while getting selected products for cancellation: %w", err)
		return err
	}

	// Удаляем каждый товар
	for productID := range selected {
		err = s.redisStore.RemoveProductFromCart(ctx, userID, productID)
		if err != nil {
			s.sugarLogger.Errorf("error while removing product %d from cart: %w", productID, err)
			// Продолжаем удаление других товаров
		}
	}

	s.sugarLogger.Infof("Order cancelled for user %d", userID)
	return nil
}
