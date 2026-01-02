package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	PGUser                 string
	PGPassword             string
	PGName                 string
	PGHost                 string
	PGPort                 string
	HTTPPort               int
	RedisAddr              string
	RedisPassword          string
	RedisDB                int
	GRPCJWTClientPort      string
	RateLimitRPS           int
	GRPCProductsClientPort string
	GRPCSagaClientPort     string
	GRPCOrderClientPort    string
	KafkaBroker            string
	KafkaGroup             string
	KafkaTopic             string
	MaxProductQuantity     int
}

func MustLoad() (*Config, error) {
	const op = "config.MustLoad"
	//nolint:errcheck // .env файл опционален, игнорируем ошибку если его нет
	_ = godotenv.Load(".env")
	HTTPPort, err := strconv.Atoi(os.Getenv("HTTP_PORT"))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	RedisDB, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		RedisDB = 0 // default
	}

	RateLimitRPS, err := strconv.Atoi(os.Getenv("RATE_LIMIT_RPS"))
	if err != nil {
		RateLimitRPS = 60 // default 60 requests per minute
	}

	MaxProductQuantity, err := strconv.Atoi(os.Getenv("MAX_PRODUCT_QUANTITY"))
	if err != nil {
		MaxProductQuantity = 100 // default max 100 items per product
	}

	return &Config{
		PGUser:                 os.Getenv("PG_USER"),
		PGPassword:             os.Getenv("PG_PASSWORD"),
		PGName:                 os.Getenv("PG_NAME"),
		PGHost:                 os.Getenv("PG_HOST"),
		PGPort:                 os.Getenv("PG_PORT"),
		HTTPPort:               HTTPPort,
		RedisAddr:              os.Getenv("REDIS_ADDR"),
		RedisPassword:          os.Getenv("REDIS_PASSWORD"),
		RedisDB:                RedisDB,
		GRPCJWTClientPort:      os.Getenv("GRPC_JWT_CLIENT_PORT"),
		RateLimitRPS:           RateLimitRPS,
		GRPCProductsClientPort: os.Getenv("GRPC_PRODUCTS_CLIENT_PORT"),
		GRPCOrderClientPort:    os.Getenv("GRPC_ORDER_CLIENT_PORT"),
		GRPCSagaClientPort:     os.Getenv("GRPC_SAGA_CLIENT_PORT"),
		KafkaBroker:            os.Getenv("KAFKA_BROKER"),
		KafkaGroup:             os.Getenv("KAFKA_GROUP_ID"),
		KafkaTopic:             os.Getenv("KAFKA_TOPIC"),
		MaxProductQuantity:     MaxProductQuantity,
	}, nil
}
