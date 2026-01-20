package validator

import (
	"testing"
	"time"

	"github.com/vsespontanno/eCommerce/services/sso-service/internal/domain/models"
	"github.com/vsespontanno/eCommerce/services/sso-service/internal/lib/jwt"
)

func TestValidateToken(t *testing.T) {
	jwtSecret := "test-secret"
	service := New(jwtSecret)

	user := models.User{
		ID:    1,
		Email: "test@example.com",
	}
	token, err := jwt.NewToken(user, jwtSecret, time.Hour)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	valid, userID, err := service.ValidateJWTToken(token)
	if err != nil {
		t.Errorf("Error validating token: %v", err)
	}
	if !valid {
		t.Error("Token should be valid")
	}
	if userID != float64(user.ID) {
		t.Errorf("Expected userID %v, got %v", user.ID, userID)
	}

	wrongSecretToken, _ := jwt.NewToken(user, "wrong-secret", time.Hour)
	valid, _, err = service.ValidateJWTToken(wrongSecretToken)
	if err == nil {
		t.Error("Expected error for token signed with wrong secret")
	}
	if valid {
		t.Error("Token signed with wrong secret should not be valid")
	}

	expiredToken, _ := jwt.NewToken(user, jwtSecret, -time.Hour)
	valid, _, err = service.ValidateJWTToken(expiredToken)
	if err == nil {
		t.Error("Expected error for expired token")
	}
	if valid {
		t.Error("Expired token should not be valid")
	}
}
