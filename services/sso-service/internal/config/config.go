package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	JWTSecret string
	GRPCPort  int
	HTTPPort  int
	User      string
	Password  string
	Name      string
	Host      string
	PGPort    string
}

func MustLoad() (*Config, error) {
	const op = "config.MustLoad"

	// Пытаемся загрузить .env файл, но не падаем если его нет
	// В Kubernetes переменные окружения уже установлены через ConfigMap/Secret
	//nolint:errcheck // .env файл опционален, игнорируем ошибку если его нет
	_ = godotenv.Load(".env")

	grpcPort, err := strconv.Atoi(os.Getenv("GRPC_PORT"))
	if err != nil {
		return nil, fmt.Errorf("%s: failed to parse GRPC_PORT: %w", op, err)
	}

	// HTTP порт по умолчанию 8080 если не указан
	httpPort := 8080
	if httpPortStr := os.Getenv("HTTP_PORT"); httpPortStr != "" {
		httpPort, err = strconv.Atoi(httpPortStr)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to parse HTTP_PORT: %w", op, err)
		}
	}

	cfg := &Config{
		JWTSecret: os.Getenv("JWT_SECRET"),
		GRPCPort:  grpcPort,
		HTTPPort:  httpPort,
		User:      os.Getenv("PG_USER"),
		Password:  os.Getenv("PG_PASSWORD"),
		Name:      os.Getenv("PG_NAME"),
		Host:      os.Getenv("PG_HOST"),
		PGPort:    os.Getenv("PG_PORT"),
	}

	// Валидация обязательных полей
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("%s: JWT_SECRET is required", op)
	}
	if cfg.User == "" {
		return nil, fmt.Errorf("%s: PG_USER is required", op)
	}
	if cfg.Password == "" {
		return nil, fmt.Errorf("%s: PG_PASSWORD is required", op)
	}
	if cfg.Name == "" {
		return nil, fmt.Errorf("%s: PG_NAME is required", op)
	}
	if cfg.Host == "" {
		return nil, fmt.Errorf("%s: PG_HOST is required", op)
	}
	if cfg.PGPort == "" {
		return nil, fmt.Errorf("%s: PG_PORT is required", op)
	}

	return cfg, nil
}
