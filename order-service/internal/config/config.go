package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	GRPCServerPort         int
	GRPCWalletClientPort   string
	HTTPProductsClientPort string
	KafkaHost              string
	PGUser                 string
	PGPassword             string
	PGName                 string
	PGHost                 string
	PGPort                 string
}

func MustLoad() (*Config, error) {
	var cfg Config
	if err := godotenv.Load(); err != nil {
		return nil, err
	}
	cfg.GRPCServerPort = getEnvAsInt("GRPC_SERVER_PORT", 50051)
	cfg.GRPCWalletClientPort = os.Getenv("GRPC_WALLET_CLIENT_PORT")
	cfg.HTTPProductsClientPort = os.Getenv("HTTP_PRODUCTS_CLIENT_PORT")
	cfg.KafkaHost = os.Getenv("KAFKA_HOST")
	cfg.PGUser = os.Getenv("PG_USER")
	cfg.PGPassword = os.Getenv("PG_PASSWORD")
	cfg.PGName = os.Getenv("PG_NAME")
	cfg.PGHost = os.Getenv("PG_HOST")
	cfg.PGPort = os.Getenv("PG_PORT")
	return &cfg, nil
}

func getEnvAsInt(name string, defaultVal int) int {
	val := os.Getenv(name)
	if val == "" {
		return defaultVal
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return intVal
}
