package repository

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func ConnectToRedis(addr string, password string) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0, // use default DB
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}
	return rdb
}
