package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/vsespontanno/eCommerce/cart-service/internal/domain/models"
)

type OrderStore struct {
	rdb *redis.Client
}

func NewOrderStore(rdb *redis.Client) *OrderStore {
	return &OrderStore{
		rdb: rdb,
	}
}

func (s *OrderStore) AddAllQuantityOfProductToCart(ctx context.Context, userID int64, productID int64, quantity int64) error {
	_, err := s.rdb.HSet(ctx, "cart:"+strconv.FormatInt(userID, 10), strconv.FormatInt(productID, 10), strconv.FormatInt(quantity, 10)).Result()
	if err != nil {
		return err
	}
	return nil
}

func (s *OrderStore) AddAllProductsToCart(ctx context.Context, userID int64, productIDs []int64) error {
	for _, productID := range productIDs {
		err := s.AddAllQuantityOfProductToCart(ctx, userID, productID, 1)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *OrderStore) AddToCart(ctx context.Context, userID int64, productID int64) error {
	key := "cart:" + strconv.FormatInt(userID, 10)
	field := strconv.FormatInt(productID, 10)
	_, err := s.rdb.HIncrBy(ctx, key, field, 1).Result()
	if err != nil {
		return err
	}
	return s.rdb.Expire(ctx, key, 30*24*time.Hour).Err()
}

func (s *OrderStore) RemoveOneFromCart(ctx context.Context, userID int64, productID int64) error {
	key := "cart:" + strconv.FormatInt(userID, 10)
	field := strconv.FormatInt(productID, 10)
	newQuantity, err := s.rdb.HIncrBy(ctx, key, field, -1).Result()
	if err != nil {
		return err
	}
	if newQuantity <= 0 {
		_, err = s.rdb.HDel(ctx, key, field).Result()
	}
	return err

}

func (s *OrderStore) RemoveProductFromCart(ctx context.Context, userID int64, productID int64) error {
	_, err := s.rdb.HDel(ctx, "cart:"+strconv.FormatInt(userID, 10), strconv.FormatInt(productID, 10)).Result()
	if err != nil {
		return err
	}
	return nil
}

func (s *OrderStore) GetCart(ctx context.Context, userID int64) (map[int64]int64, error) {
	key := "cart:" + strconv.FormatInt(userID, 10)
	items, err := s.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	result := make(map[int64]int64)
	for pidStr, qtyStr := range items {
		pid, err := strconv.ParseInt(pidStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid product ID: %w", err)
		}
		qty, err := strconv.ParseInt(qtyStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid quantity: %w", err)
		}
		result[pid] = qty
	}
	return result, nil
}

func (s *OrderStore) GetProductQuantity(ctx context.Context, userID int64, productID int64) (int64, error) {
	key := "cart:" + strconv.FormatInt(userID, 10)
	field := strconv.FormatInt(productID, 10)
	q, err := s.rdb.HGet(ctx, key, field).Int64()
	if err != nil {
		if err == redis.Nil {
			return 0, models.ErrProductIsNotInCart
		}
		return 0, err
	}
	return q, err
}

// To weird cause user never writes quantity; he only uses the button "+" or "-" so default value would be 1
// func (s *OrderStore) RemoveFromCart(ctx context.Context, userID int64, productID int64, quantity int64) error {
// 	if quantity == 0 {
// 		_, err := s.rdb.HDel(ctx, "cart:"+strconv.FormatInt(userID, 10), strconv.FormatInt(productID, 10)).Result()
// 		if err != nil {
// 			return err
// 		}
// 		return nil
// 	}

// 	currentQuantityStr, err := s.rdb.HGet(ctx, "cart:"+strconv.FormatInt(userID, 10), strconv.FormatInt(productID, 10)).Result()
// 	if err == redis.Nil {
// 		return nil // Product not in cart, nothing to remove
// 	} else if err != nil {
// 		return err
// 	}

// 	currentQuantity, err := strconv.ParseInt(currentQuantityStr, 10, 64)
// 	if err != nil {
// 		return err
// 	}

// 	newQuantity := currentQuantity - quantity
// 	if newQuantity <= 0 {
// 		_, err := s.rdb.HDel(ctx, "cart:"+strconv.FormatInt(userID, 10), strconv.FormatInt(productID, 10)).Result()
// 		if err != nil {
// 			return err
// 		}
// 	} else {
// 		_, err := s.rdb.HSet(ctx, "cart:"+strconv.FormatInt(userID, 10), strconv.FormatInt(productID, 10), newQuantity).Result()
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }
