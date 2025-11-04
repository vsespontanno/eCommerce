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
	User      string
	Password  string
	Name      string
	Host      string
	PGPort    string
}

func MustLoad() (*Config, error) {
	const op = "config.MustLoad"
	err := godotenv.Load(".env")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	grpcPort, err := strconv.Atoi(os.Getenv("GRPC_PORT"))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Config{
		JWTSecret: os.Getenv("JWT_SECRET"),
		GRPCPort:  grpcPort,
		User:      os.Getenv("PG_USER"),
		Password:  os.Getenv("PG_PASSWORD"),
		Name:      os.Getenv("PG_NAME"),
		Host:      os.Getenv("PG_HOST"),
		PGPort:    os.Getenv("PG_PORT"),
	}, nil
}
