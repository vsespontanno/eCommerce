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
}

func MustLoad() (*Config, error) {
	const op = "config.MustLoad"
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
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
	}, nil
}
