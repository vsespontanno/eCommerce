package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/vsespontanno/eCommerce/services/cart-service/internal/domain/apperrors"
	"github.com/vsespontanno/eCommerce/services/cart-service/internal/domain/cart/entity"
	"go.uber.org/zap"
)

type CartStore struct {
	rdb    *redis.Client
	logger *zap.SugaredLogger
}

func NewCartStore(rdb *redis.Client, logger *zap.SugaredLogger) *CartStore {
	return &CartStore{
		rdb:    rdb,
		logger: logger,
	}
}

func (s *CartStore) IncrementInCart(ctx context.Context, userID int64, productID int64) error {
	key := fmt.Sprintf("cart:%d", userID)
	field := strconv.FormatInt(productID, 10)
	existingJSON, err := s.rdb.HGet(ctx, key, field).Result()
	if err != nil && err != redis.Nil {
		s.logger.Errorw("Failed to add product to cart", "error", err, "stage", "AddToCart")
		return err
	}
	if existingJSON != "" {
		var existing entity.CartItem
		if err := json.Unmarshal([]byte(existingJSON), &existing); err != nil {
			s.logger.Errorw("Failed to add product to cart", "error", err, "stage", "AddToCart")
			return err
		}
		existing.Quantity += 1
		data, err := json.Marshal(existing)
		if err != nil {
			s.logger.Errorw("Failed to add product to cart", "error", err, "stage", "AddToCart")
			return err
		}

		if _, err := s.rdb.HSet(ctx, key, field, data).Result(); err != nil {
			s.logger.Errorw("Failed to add product to cart", "error", err, "stage", "AddToCart")
			return err
		}
		return s.rdb.Expire(ctx, key, 24*time.Hour).Err()
	}
	return apperrors.ErrProductIsNotInCart
}

func (s *CartStore) AddNewProductToCart(ctx context.Context, userID int64, product *entity.CartItem) error {
	key := fmt.Sprintf("cart:%d", userID)
	field := strconv.FormatInt(product.ProductID, 10)
	data, err := json.Marshal(product)
	if err != nil {
		s.logger.Errorw("Failed to add product to cart", "error", err, "stage", "AddToCart")
		return err
	}
	if _, err := s.rdb.HSet(ctx, key, field, data).Result(); err != nil {
		s.logger.Errorw("Failed to add product to cart", "error", err, "stage", "AddToCart")
		return err
	}
	return s.rdb.Expire(ctx, key, 30*24*time.Hour).Err()
}

func (s *CartStore) SaveCart(ctx context.Context, userID int64, cart *entity.Cart) error {
	key := fmt.Sprintf("cart:%d", userID)
	for _, item := range cart.Items {
		data, err := json.Marshal(item)
		if err != nil {
			s.logger.Errorw("Failed to add product to cart", "error", err, "stage", "AddToCart")
			return err
		}
		field := strconv.FormatInt(item.ProductID, 10)
		if _, err := s.rdb.HSet(ctx, key, field, data).Result(); err != nil {
			s.logger.Errorw("Failed to add product to cart", "error", err, "stage", "AddToCart")
			return err
		}
	}
	return s.rdb.Expire(ctx, key, 24*time.Hour).Err()
}

func (s *CartStore) DecrementInCart(ctx context.Context, userID, productID int64) error {
	key := fmt.Sprintf("cart:%d", userID)
	field := strconv.FormatInt(productID, 10)

	jsonStr, err := s.rdb.HGet(ctx, key, field).Result()
	if err == redis.Nil {
		return apperrors.ErrProductIsNotInCart
	}
	if err != nil {
		s.logger.Errorw("Failed to get product for removal", "error", err)
		return err
	}

	var p entity.CartItem
	if unmarshalErr := json.Unmarshal([]byte(jsonStr), &p); unmarshalErr != nil {
		s.logger.Errorw("Failed to unmarshal product", "error", unmarshalErr)
		return unmarshalErr
	}

	p.Quantity--
	if p.Quantity <= 0 {
		_, err = s.rdb.HDel(ctx, key, field).Result()
		return err
	}

	data, marshalErr := json.Marshal(p)
	if marshalErr != nil {
		s.logger.Errorw("Failed to marshal product", "error", marshalErr)
		return marshalErr
	}
	_, err = s.rdb.HSet(ctx, key, field, data).Result()
	return err
}

func (s *CartStore) RemoveProductFromCart(ctx context.Context, userID int64, productID int64) error {
	_, err := s.rdb.HDel(ctx, "cart:"+strconv.FormatInt(userID, 10), strconv.FormatInt(productID, 10)).Result()
	if err != nil {
		s.logger.Errorw("Failed to remove product from cart", "error", err, "stage", "RemoveProductFromCart")
		return err
	}
	return nil
}

func (s *CartStore) GetCartProducts(ctx context.Context, userID int64) (*entity.Cart, error) {
	key := fmt.Sprintf("cart:%d", userID)
	items, err := s.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		s.logger.Errorw("Failed to get cart", "error", err, "stage", "GetCart")
		return nil, err
	}

	var cart entity.Cart
	for _, jsonStr := range items {
		var item entity.CartItem
		if err := json.Unmarshal([]byte(jsonStr), &item); err != nil {
			s.logger.Errorw("Failed to unmarshal product", "error", err, "stage", "GetCart")
			return nil, err
		}
		cart.Items = append(cart.Items, item)
	}
	return &cart, nil
}

func (s *CartStore) ClearCart(ctx context.Context, userID int64) error {
	_, err := s.rdb.Del(ctx, "cart:"+strconv.FormatInt(userID, 10)).Result()
	if err != nil {
		s.logger.Errorw("Failed to clear cart", "error", err, "stage", "ClearCart")
		return err
	}
	return nil
}

func (s *CartStore) GetCart(ctx context.Context, userID int64) (*entity.Cart, error) {
	key := fmt.Sprintf("cart:%d", userID)
	items, err := s.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		s.logger.Errorw("Failed to get cart", "error", err, "stage", "GetCart")
		return &entity.Cart{}, err
	}
	if len(items) == 0 {
		return &entity.Cart{}, apperrors.ErrNoCartFound

	}

	var cart entity.Cart
	for _, jsonStr := range items {
		var item entity.CartItem
		if err := json.Unmarshal([]byte(jsonStr), &item); err != nil {
			s.logger.Errorw("Failed to unmarshal product", "error", err, "stage", "GetCart")
			return nil, err
		}
		cart.Items = append(cart.Items, item)
	}

	return &cart, nil
}

func (s *CartStore) GetProduct(ctx context.Context, userID, productID int64) (*entity.CartItem, error) {
	key := fmt.Sprintf("cart:%d", userID)
	field := strconv.FormatInt(productID, 10)

	jsonStr, err := s.rdb.HGet(ctx, key, field).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, apperrors.ErrProductIsNotInCart
		}
		s.logger.Errorw("Failed to get product from cart", "error", err, "stage", "GetProduct")
		return nil, err
	}

	var p entity.CartItem
	if err := json.Unmarshal([]byte(jsonStr), &p); err != nil {
		s.logger.Errorw("Failed to unmarshal product", "error", err, "stage", "GetProduct")
		return nil, err
	}

	return &p, nil
}

func (s *CartStore) DeleteProduct(ctx context.Context, userID, productID int64) error {
	_, err := s.rdb.HDel(ctx, "cart:"+strconv.FormatInt(userID, 10), strconv.FormatInt(productID, 10)).Result()
	if err != nil {
		s.logger.Errorw("Failed to remove product from cart", "error", err, "stage", "RemoveProductFromCart")
		return err
	}
	return nil
}
