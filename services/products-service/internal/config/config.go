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
	GRPCProductsServerPort int
	GRPCSagaServerPort     int
	GRPCJwtPort            string
}

func MustLoad() (*Config, error) {
	const op = "config.MustLoad"
	//nolint:errcheck // .env файл опционален, игнорируем ошибку если его нет
	_ = godotenv.Load(".env")
	HTTPPort, err := strconv.Atoi(os.Getenv("HTTP_PORT"))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	GRPCProductsServerPort, err := strconv.Atoi(os.Getenv("GRPC_PRODUCTS_SERVER_PORT"))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	GRPCSagaServerPort, err := strconv.Atoi(os.Getenv("GRPC_SAGA_SERVER_PORT"))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Config{
		PGUser:                 os.Getenv("PG_USER"),
		PGPassword:             os.Getenv("PG_PASSWORD"),
		PGName:                 os.Getenv("PG_NAME"),
		PGHost:                 os.Getenv("PG_HOST"),
		PGPort:                 os.Getenv("PG_PORT"),
		HTTPPort:               HTTPPort,
		GRPCProductsServerPort: GRPCProductsServerPort,
		GRPCSagaServerPort:     GRPCSagaServerPort,
		GRPCJwtPort:            os.Getenv("GRPC_JWT_CLIENT_PORT"),
	}, nil
}
