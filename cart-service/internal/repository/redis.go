package repository

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func ConnectToRedis(addr string) *redis.Client {
	host := "localhost:" + addr
	rdb := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: "",
		DB:       0, // use default DB
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}
	return rdb
}
