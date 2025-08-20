package jwt

import (
	"testing"
	"time"

	"github.com/vsespontanno/eCommerce/sso-service/internal/domain/models"
)

func TestNewToken(t *testing.T) {
	user := models.User{
		ID:    1,
		Email: "test@example.com",
	}
	app := models.App{
		ID:     1,
		Secret: "supersecret",
	}
	token, err := NewToken(user, app, time.Hour)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}
	if token == "" {
		t.Fatal("Expected token to be non-empty")
	}
}
