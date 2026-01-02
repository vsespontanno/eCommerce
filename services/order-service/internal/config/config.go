package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	PGUser         string
	PGPassword     string
	PGName         string
	PGHost         string
	PGPort         string
	GRPCServerPort int
}

func MustLoad() (*Config, error) {
	const op = "config.MustLoad"
	//nolint:errcheck // .env файл опционален, игнорируем ошибку если его нет
	_ = godotenv.Load(".env")
	GRPCServerPort, err := strconv.Atoi(os.Getenv("GRPC_SERVER_PORT"))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Config{
		PGUser:         os.Getenv("PG_USER"),
		PGPassword:     os.Getenv("PG_PASSWORD"),
		PGName:         os.Getenv("PG_NAME"),
		PGHost:         os.Getenv("PG_HOST"),
		PGPort:         os.Getenv("PG_PORT"),
		GRPCServerPort: GRPCServerPort,
	}, nil
}
