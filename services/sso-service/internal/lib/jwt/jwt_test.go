package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vsespontanno/eCommerce/services/sso-service/internal/domain/models"
)

func TestNewToken(t *testing.T) {
	user := models.User{
		ID:    1,
		Email: "test@example.com",
	}
	secret := "test-secret"
	duration := time.Hour

	token, err := NewToken(user, secret, duration)
	if err != nil {
		t.Fatalf("NewToken() error = %v", err)
	}

	if token == "" {
		t.Error("NewToken() returned empty token")
	}

	parsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("Failed to get claims")
	}

	if int64(claims["uid"].(float64)) != user.ID {
		t.Errorf("uid = %v, want %v", claims["uid"], user.ID)
	}

	if claims["email"] != user.Email {
		t.Errorf("email = %v, want %v", claims["email"], user.Email)
	}
}

func TestNewTokenWithExpiry(t *testing.T) {
	user := models.User{
		ID:    1,
		Email: "test@example.com",
	}
	secret := "test-secret"
	duration := time.Hour

	tokenPair, err := NewTokenWithExpiry(user, secret, duration)
	if err != nil {
		t.Fatalf("NewTokenWithExpiry() error = %v", err)
	}

	if tokenPair.Token == "" {
		t.Error("NewTokenWithExpiry() returned empty token")
	}

	if tokenPair.ExpiresAt == 0 {
		t.Error("NewTokenWithExpiry() returned zero ExpiresAt")
	}

	expectedExpiry := time.Now().Add(duration).Unix()
	if tokenPair.ExpiresAt < expectedExpiry-5 || tokenPair.ExpiresAt > expectedExpiry+5 {
		t.Errorf("ExpiresAt = %v, want approximately %v", tokenPair.ExpiresAt, expectedExpiry)
	}

	parsed, err := jwt.Parse(tokenPair.Token, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("Failed to get claims")
	}

	if int64(claims["uid"].(float64)) != user.ID {
		t.Errorf("uid = %v, want %v", claims["uid"], user.ID)
	}

	if claims["email"] != user.Email {
		t.Errorf("email = %v, want %v", claims["email"], user.Email)
	}
}

func TestNewToken_InvalidSecret(t *testing.T) {
	user := models.User{
		ID:    1,
		Email: "test@example.com",
	}
	secret := "test-secret"
	wrongSecret := "wrong-secret"
	duration := time.Hour

	token, err := NewToken(user, secret, duration)
	if err != nil {
		t.Fatalf("NewToken() error = %v", err)
	}

	_, err = jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(wrongSecret), nil
	})
	if err == nil {
		t.Error("Expected error when parsing with wrong secret")
	}
}
