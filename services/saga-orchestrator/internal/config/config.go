package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	GRPCServerPort         int
	HTTPHealthPort         int
	GRPCWalletClientPort   string
	GRPCProductsClientPort string
	KafkaBroker            string
	KafkaGroup             string
	KafkaTopic             string
	PGUser                 string
	PGPassword             string
	PGName                 string
	PGHost                 string
	PGPort                 string
}

func MustLoad() (*Config, error) {
	var cfg Config
	//nolint:errcheck // .env файл опционален, игнорируем ошибку если его нет
	_ = godotenv.Load(".env")
	cfg.GRPCServerPort = getEnvAsInt("GRPC_SERVER_PORT", 50051)
	cfg.HTTPHealthPort = getEnvAsInt("HTTP_HEALTH_PORT", 8080)
	cfg.GRPCWalletClientPort = os.Getenv("GRPC_WALLET_CLIENT_PORT")
	cfg.GRPCProductsClientPort = os.Getenv("GRPC_PRODUCTS_CLIENT_PORT")
	cfg.KafkaBroker = os.Getenv("KAFKA_BROKER")
	cfg.KafkaGroup = os.Getenv("KAFKA_GROUP_ID")
	cfg.KafkaTopic = os.Getenv("KAFKA_TOPIC")
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
