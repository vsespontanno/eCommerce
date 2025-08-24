package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	GRPCServer int
	GRPCClient string
	PGPort     string
	Host       string
	User       string
	Password   string
	Name       string
}

func MustLoad() (*Config, error) {
	var cfg Config
	if err := godotenv.Load(); err != nil {
		return nil, err
	}
	cfg.GRPCServer = getEnvAsInt("GRPC_SERVER_PORT", 50051)
	cfg.GRPCClient = os.Getenv("GRPC_CLIENT_PORT")
	cfg.PGPort = os.Getenv("PG_PORT")
	cfg.Host = os.Getenv("PG_HOST")
	cfg.User = os.Getenv("PG_USER")
	cfg.Password = os.Getenv("PG_PASSWORD")
	cfg.Name = os.Getenv("PG_NAME")
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
