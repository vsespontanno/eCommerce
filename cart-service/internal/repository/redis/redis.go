package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/vsespontanno/eCommerce/cart-service/internal/domain/models"
	"go.uber.org/zap"
)

type Producter interface {
	GetProduct(ctx context.Context, productID int64) (*models.Product, error)
}

type OrderStore struct {
	rdb           *redis.Client
	logger        *zap.SugaredLogger
	productClient Producter
}

func NewOrderStore(rdb *redis.Client, logger *zap.SugaredLogger) *OrderStore {
	return &OrderStore{
		rdb:    rdb,
		logger: logger,
	}
}

func (s *OrderStore) AddToCart(ctx context.Context, userID int64, productID int64) error {
	key := fmt.Sprintf("cart:%d", userID)
	field := strconv.FormatInt(productID, 10)
	existingJSON, err := s.rdb.HGet(ctx, key, field).Result()
	if err != nil && err != redis.Nil {
		s.logger.Errorw("Failed to add product to cart", "error", err, "stage", "AddToCart")
		return err
	}
	if existingJSON != "" {
		var existing models.Product
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
		return s.rdb.Expire(ctx, key, 30*24*time.Hour).Err()
	}

	newProduct, err := s.productClient.GetProduct(ctx, productID)
	if err != nil {
		s.logger.Errorw("Failed to add product to cart", "error", err, "stage", "AddToCart")
		return err
	}
	data, err := json.Marshal(newProduct)
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

func (s *OrderStore) RemoveOneFromCart(ctx context.Context, userID, productID int64) error {
	key := fmt.Sprintf("cart:%d", userID)
	field := strconv.FormatInt(productID, 10)

	jsonStr, err := s.rdb.HGet(ctx, key, field).Result()
	if err == redis.Nil {
		return models.ErrProductIsNotInCart
	}
	if err != nil {
		s.logger.Errorw("Failed to get product for removal", "error", err)
		return err
	}

	var p models.Product
	if err := json.Unmarshal([]byte(jsonStr), &p); err != nil {
		s.logger.Errorw("Failed to unmarshal product", "error", err)
		return err
	}

	p.Quantity--
	if p.Quantity <= 0 {
		_, err = s.rdb.HDel(ctx, key, field).Result()
		return err
	}

	data, _ := json.Marshal(p)
	_, err = s.rdb.HSet(ctx, key, field, data).Result()
	return err
}

func (s *OrderStore) RemoveProductFromCart(ctx context.Context, userID int64, productID int64) error {
	_, err := s.rdb.HDel(ctx, "cart:"+strconv.FormatInt(userID, 10), strconv.FormatInt(productID, 10)).Result()
	if err != nil {
		s.logger.Errorw("Failed to remove product from cart", "error", err, "stage", "RemoveProductFromCart")
		return err
	}
	return nil
}

func (s *OrderStore) GetCart(ctx context.Context, userID int64) ([]models.Product, error) {
	key := fmt.Sprintf("cart:%d", userID)
	items, err := s.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		s.logger.Errorw("Failed to get cart", "error", err, "stage", "GetCart")
		return nil, err
	}

	var products []models.Product
	for _, jsonStr := range items {
		var p models.Product
		if err := json.Unmarshal([]byte(jsonStr), &p); err != nil {
			s.logger.Errorw("Failed to unmarshal product", "error", err, "stage", "GetCart")
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

func (s *OrderStore) GetProduct(ctx context.Context, userID, productID int64) (*models.Product, error) {
	key := fmt.Sprintf("cart:%d", userID)
	field := strconv.FormatInt(productID, 10)

	jsonStr, err := s.rdb.HGet(ctx, key, field).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, models.ErrProductIsNotInCart
		}
		s.logger.Errorw("Failed to get product from cart", "error", err, "stage", "GetProduct")
		return nil, err
	}

	var p models.Product
	if err := json.Unmarshal([]byte(jsonStr), &p); err != nil {
		s.logger.Errorw("Failed to unmarshal product", "error", err, "stage", "GetProduct")
		return nil, err
	}

	return &p, nil
}
