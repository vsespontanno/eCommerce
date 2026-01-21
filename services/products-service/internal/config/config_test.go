package config

import (
	"os"
	"testing"
)

func TestMustLoad(t *testing.T) {
	// Helper function to set env vars
	setEnv := func(vars map[string]string) {
		for k, v := range vars {
			os.Setenv(k, v)
		}
	}

	// Helper function to unset env vars
	unsetEnv := func(vars []string) {
		for _, k := range vars {
			os.Unsetenv(k)
		}
	}

	envVars := []string{
		"HTTP_PORT",
		"GRPC_PRODUCTS_SERVER_PORT",
		"GRPC_SAGA_SERVER_PORT",
		"PG_USER",
		"PG_PASSWORD",
		"PG_NAME",
		"PG_HOST",
		"PG_PORT",
		"GRPC_JWT_CLIENT_PORT",
	}

	t.Run("Success", func(t *testing.T) {
		setEnv(map[string]string{
			"HTTP_PORT":                 "8080",
			"GRPC_PRODUCTS_SERVER_PORT": "50051",
			"GRPC_SAGA_SERVER_PORT":     "50052",
			"PG_USER":                   "user",
			"PG_PASSWORD":               "pass",
			"PG_NAME":                   "db",
			"PG_HOST":                   "localhost",
			"PG_PORT":                   "5432",
			"GRPC_JWT_CLIENT_PORT":      "50053",
		})
		defer unsetEnv(envVars)

		cfg, err := MustLoad()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if cfg.HTTPPort != 8080 {
			t.Errorf("Expected HTTPPort 8080, got %d", cfg.HTTPPort)
		}
		if cfg.GRPCProductsServerPort != 50051 {
			t.Errorf("Expected GRPCProductsServerPort 50051, got %d", cfg.GRPCProductsServerPort)
		}
		if cfg.GRPCSagaServerPort != 50052 {
			t.Errorf("Expected GRPCSagaServerPort 50052, got %d", cfg.GRPCSagaServerPort)
		}
		if cfg.PGUser != "user" {
			t.Errorf("Expected PGUser user, got %s", cfg.PGUser)
		}
	})

	t.Run("Missing HTTP_PORT", func(t *testing.T) {
		unsetEnv(envVars)
		setEnv(map[string]string{
			"GRPC_PRODUCTS_SERVER_PORT": "50051",
			"GRPC_SAGA_SERVER_PORT":     "50052",
		})
		defer unsetEnv(envVars)

		_, err := MustLoad()
		if err == nil {
			t.Error("Expected error for missing HTTP_PORT")
		}
	})

	t.Run("Invalid HTTP_PORT", func(t *testing.T) {
		unsetEnv(envVars)
		setEnv(map[string]string{
			"HTTP_PORT":                 "invalid",
			"GRPC_PRODUCTS_SERVER_PORT": "50051",
			"GRPC_SAGA_SERVER_PORT":     "50052",
		})
		defer unsetEnv(envVars)

		_, err := MustLoad()
		if err == nil {
			t.Error("Expected error for invalid HTTP_PORT")
		}
	})

	t.Run("Missing GRPC_PRODUCTS_SERVER_PORT", func(t *testing.T) {
		unsetEnv(envVars)
		setEnv(map[string]string{
			"HTTP_PORT":             "8080",
			"GRPC_SAGA_SERVER_PORT": "50052",
		})
		defer unsetEnv(envVars)

		_, err := MustLoad()
		if err == nil {
			t.Error("Expected error for missing GRPC_PRODUCTS_SERVER_PORT")
		}
	})

	t.Run("Missing GRPC_SAGA_SERVER_PORT", func(t *testing.T) {
		unsetEnv(envVars)
		setEnv(map[string]string{
			"HTTP_PORT":                 "8080",
			"GRPC_PRODUCTS_SERVER_PORT": "50051",
		})
		defer unsetEnv(envVars)

		_, err := MustLoad()
		if err == nil {
			t.Error("Expected error for missing GRPC_SAGA_SERVER_PORT")
		}
	})
}
