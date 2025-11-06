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
	HTTPPort       int
	GRPCServerPort int
	GRPCJwtPort    string
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
	GRPCPort, err := strconv.Atoi(os.Getenv("GRPC_SERVER_PORT"))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Config{
		PGUser:         os.Getenv("PG_USER"),
		PGPassword:     os.Getenv("PG_PASSWORD"),
		PGName:         os.Getenv("PG_NAME"),
		PGHost:         os.Getenv("PG_HOST"),
		PGPort:         os.Getenv("PG_PORT"),
		HTTPPort:       HTTPPort,
		GRPCServerPort: GRPCPort,
		GRPCJwtPort:    os.Getenv("GRPC_JWT_CLIENT_PORT"),
	}, nil
}
