package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/vsespontanno/eCommerce/cart-service/internal/domain/models"
	"go.uber.org/zap"
)

type Cleaner struct {
	client *redis.Client
	logger *zap.SugaredLogger
}

func NewCleaner(client *redis.Client, logger *zap.SugaredLogger) *Cleaner {
	return &Cleaner{client: client, logger: logger}
}

func (c *Cleaner) CleanCart(ctx context.Context, order *models.OrderEvent) error {
	key := fmt.Sprintf("cart:%d", order.UserID)
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		c.logger.Errorw("Failed to delete redis cart", "userID", order.UserID, "error", err)
		return err
	}
	c.logger.Infow("Redis cart cleared", "userID", order.UserID)
	return nil
}
