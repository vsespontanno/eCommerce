package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMustLoad(t *testing.T) {
	// Save original env vars
	originalHTTPPort := os.Getenv("HTTP_PORT")
	originalRedisDB := os.Getenv("REDIS_DB")
	originalRateLimitRPS := os.Getenv("RATE_LIMIT_RPS")
	originalMaxProductQuantity := os.Getenv("MAX_PRODUCT_QUANTITY")

	defer func() {
		// Restore env vars
		os.Setenv("HTTP_PORT", originalHTTPPort)
		os.Setenv("REDIS_DB", originalRedisDB)
		os.Setenv("RATE_LIMIT_RPS", originalRateLimitRPS)
		os.Setenv("MAX_PRODUCT_QUANTITY", originalMaxProductQuantity)
	}()

	t.Run("Success", func(t *testing.T) {
		os.Setenv("HTTP_PORT", "8080")
		os.Setenv("REDIS_DB", "1")
		os.Setenv("RATE_LIMIT_RPS", "100")
		os.Setenv("MAX_PRODUCT_QUANTITY", "50")

		cfg, err := MustLoad()
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, 8080, cfg.HTTPPort)
		assert.Equal(t, 1, cfg.RedisDB)
		assert.Equal(t, 100, cfg.RateLimitRPS)
		assert.Equal(t, 50, cfg.MaxProductQuantity)
	})

	t.Run("Missing HTTP_PORT", func(t *testing.T) {
		os.Unsetenv("HTTP_PORT")
		cfg, err := MustLoad()
		assert.Error(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("Invalid HTTP_PORT", func(t *testing.T) {
		os.Setenv("HTTP_PORT", "invalid")
		cfg, err := MustLoad()
		assert.Error(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("Defaults", func(t *testing.T) {
		os.Setenv("HTTP_PORT", "8080")
		os.Unsetenv("REDIS_DB")
		os.Unsetenv("RATE_LIMIT_RPS")
		os.Unsetenv("MAX_PRODUCT_QUANTITY")

		cfg, err := MustLoad()
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, 0, cfg.RedisDB)
		assert.Equal(t, 60, cfg.RateLimitRPS)
		assert.Equal(t, 100, cfg.MaxProductQuantity)
	})
}
