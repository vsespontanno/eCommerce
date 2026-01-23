package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMustLoad(t *testing.T) {
	// Save original env vars
	originalGRPCServerPort := os.Getenv("GRPC_SERVER_PORT")
	originalHTTPHealthPort := os.Getenv("HTTP_HEALTH_PORT")

	defer func() {
		// Restore env vars
		os.Setenv("GRPC_SERVER_PORT", originalGRPCServerPort)
		os.Setenv("HTTP_HEALTH_PORT", originalHTTPHealthPort)
	}()

	t.Run("Success", func(t *testing.T) {
		os.Setenv("GRPC_SERVER_PORT", "50052")
		os.Setenv("HTTP_HEALTH_PORT", "8082")

		cfg, err := MustLoad()
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, 50052, cfg.GRPCServerPort)
		assert.Equal(t, 8082, cfg.HTTPHealthPort)
	})

	t.Run("Defaults", func(t *testing.T) {
		os.Unsetenv("GRPC_SERVER_PORT")
		os.Unsetenv("HTTP_HEALTH_PORT")

		cfg, err := MustLoad()
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, 50051, cfg.GRPCServerPort)
		assert.Equal(t, 8080, cfg.HTTPHealthPort)
	})

	t.Run("Invalid GRPC_SERVER_PORT", func(t *testing.T) {
		os.Setenv("GRPC_SERVER_PORT", "invalid")
		cfg, err := MustLoad()
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, 50051, cfg.GRPCServerPort) // Should fallback to default
	})

	t.Run("Invalid HTTP_HEALTH_PORT", func(t *testing.T) {
		os.Setenv("HTTP_HEALTH_PORT", "invalid")
		cfg, err := MustLoad()
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, 8080, cfg.HTTPHealthPort) // Should fallback to default
	})
}
