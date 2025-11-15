package jobs

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/vsespontanno/eCommerce/cart-service/internal/domain/models"
	"go.uber.org/zap"
)

type PgRepo interface {
	UpsertCart(ctx context.Context, userID int64, cart *[]models.CartItem) error
}

type RedisRepo interface {
	ScanKeys(ctx context.Context, pattern string) ([]string, error)
	GetCartItems(ctx context.Context, key string) ([]models.CartItem, error)
}

type CartSyncJob struct {
	pgRepo    PgRepo
	redisRepo RedisRepo
	logger    *zap.SugaredLogger
	interval  time.Duration
}

func NewCartSyncJob(pg PgRepo, rd RedisRepo, logger *zap.SugaredLogger, interval time.Duration) *CartSyncJob {
	return &CartSyncJob{
		pgRepo:    pg,
		redisRepo: rd,
		logger:    logger,
		interval:  interval,
	}
}

func (j *CartSyncJob) Start(ctx context.Context) {
	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	j.logger.Infof("CartSyncJob started (interval: %v)", j.interval)

	for {
		select {
		case <-ticker.C:
			if err := j.sync(ctx); err != nil {
				j.logger.Errorw("cart sync failed", "error", err)
			}
		case <-ctx.Done():
			j.logger.Info("CartSyncJob stopped")
			return
		}
	}
}

func (j *CartSyncJob) sync(ctx context.Context) error {
	keys, err := j.redisRepo.ScanKeys(ctx, "cart:*")
	if err != nil {
		return err
	}

	for _, key := range keys {

		// ---- 1. userID из ключа ----
		userID, err := parseUserIDFromKey(key)
		if err != nil {
			j.logger.Warnw("invalid cart key", "key", key)
			continue
		}

		// ---- 2. читаем корзину ----
		items, err := j.redisRepo.GetCartItems(ctx, key)
		if err != nil {
			j.logger.Warnw("failed to get cart from redis", "key", key, "error", err)
			continue
		}

		// ---- 3. фильтруем битые items (главное — productID = 0) ----
		valid := make([]models.CartItem, 0, len(items))
		for _, it := range items {
			if it.ProductID == 0 {
				continue
			}
			valid = append(valid, it)
		}

		// пустая корзина → ничего не пишем
		if len(valid) == 0 {
			continue
		}

		// ---- 4. пишем только валидные товары ----
		if err := j.pgRepo.UpsertCart(ctx, userID, &valid); err != nil {
			j.logger.Errorw("failed to upsert cart into Postgres",
				"userID", userID, "error", err)
		}
	}

	j.logger.Infow("CartSyncJob completed", "carts_synced", len(keys))
	return nil
}

func parseUserIDFromKey(key string) (int64, error) {
	userID := key[strings.Index(key, ":")+1:]
	return strconv.ParseInt(userID, 10, 64)
}
