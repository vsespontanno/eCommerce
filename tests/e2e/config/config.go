package config

import (
	"os"
	"time"
)

type E2EConfig struct {
	SSOBaseURL      string //8081
	WalletBaseURL   string //8080
	ProductsBaseURL string //8082
	CartBaseURL     string //8083
	OrderGRPCAddr   string

	RequestTimeout time.Duration
	PollTimeout    time.Duration
	PollInterval   time.Duration
}

func LoadConfig() *E2EConfig {
	return &E2EConfig{
		SSOBaseURL:      getEnv("SSO_BASE_URL", "http://localhost:8081"),
		WalletBaseURL:   getEnv("WALLET_BASE_URL", "http://localhost:8080"),
		ProductsBaseURL: getEnv("PRODUCTS_BASE_URL", "http://localhost:8082"),
		CartBaseURL:     getEnv("CART_BASE_URL", "http://localhost:8083"),
		OrderGRPCAddr:   getEnv("ORDER_GRPC_ADDR", "localhost:50056"),
		RequestTimeout:  10 * time.Second,
		PollTimeout:     30 * time.Second,
		PollInterval:    500 * time.Millisecond,
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
