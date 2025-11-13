package redis

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
	"github.com/vsespontanno/eCommerce/cart-service/internal/domain/models"
	"go.uber.org/zap"
)

type RedisUpdate struct {
	client *redis.Client
	logger *zap.SugaredLogger
}

func NewRedisUpdate(client *redis.Client, logger *zap.SugaredLogger) *RedisUpdate {
	return &RedisUpdate{client: client, logger: logger}
}

func (r *RedisUpdate) ScanKeys(ctx context.Context, pattern string) ([]string, error) {
	return r.client.Keys(ctx, pattern).Result()
}

func (r *RedisUpdate) GetCartItems(ctx context.Context, key string) ([]models.CartItem, error) {
	items, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var cart []models.CartItem
	for _, jsonStr := range items {
		var item models.CartItem
		if err := json.Unmarshal([]byte(jsonStr), &item); err == nil {
			cart = append(cart, item)
		}
	}

	return cart, nil
}
