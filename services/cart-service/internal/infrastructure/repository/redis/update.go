package redis

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
	"github.com/vsespontanno/eCommerce/services/cart-service/internal/domain/cart/entity"
	"go.uber.org/zap"
)

type Updater struct {
	client *redis.Client
	logger *zap.SugaredLogger
}

func NewRedisUpdater(client *redis.Client, logger *zap.SugaredLogger) *Updater {
	return &Updater{client: client, logger: logger}
}

func (r *Updater) ScanKeys(ctx context.Context, pattern string) ([]string, error) {
	return r.client.Keys(ctx, pattern).Result()
}

func (r *Updater) GetCartItems(ctx context.Context, key string) ([]entity.CartItem, error) {
	items, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var cart []entity.CartItem
	for _, jsonStr := range items {
		var item entity.CartItem
		if err := json.Unmarshal([]byte(jsonStr), &item); err == nil {
			cart = append(cart, item)
		}
	}

	return cart, nil
}
