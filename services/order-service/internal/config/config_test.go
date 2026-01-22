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
		os.Setenv("GRPC_SERVER_PORT", "50051")
		os.Setenv("HTTP_HEALTH_PORT", "8081")

		cfg, err := MustLoad()
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, 50051, cfg.GRPCServerPort)
		assert.Equal(t, 8081, cfg.HTTPHealthPort)
	})

	t.Run("Missing GRPC_SERVER_PORT", func(t *testing.T) {
		os.Unsetenv("GRPC_SERVER_PORT")
		cfg, err := MustLoad()
		assert.Error(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("Invalid GRPC_SERVER_PORT", func(t *testing.T) {
		os.Setenv("GRPC_SERVER_PORT", "invalid")
		cfg, err := MustLoad()
		assert.Error(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("Default HTTP_HEALTH_PORT", func(t *testing.T) {
		os.Setenv("GRPC_SERVER_PORT", "50051")
		os.Unsetenv("HTTP_HEALTH_PORT")

		cfg, err := MustLoad()
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, 8080, cfg.HTTPHealthPort)
	})

	t.Run("Invalid HTTP_HEALTH_PORT", func(t *testing.T) {
		os.Setenv("GRPC_SERVER_PORT", "50051")
		os.Setenv("HTTP_HEALTH_PORT", "invalid")

		cfg, err := MustLoad()
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, 8080, cfg.HTTPHealthPort) // Should fallback to default
	})
}
