package validator

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestValidateToken(t *testing.T) {
	err := godotenv.Load("../../../.env")
	if err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	service := New(jwtSecret)

	example_token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InRpbXVyQGdtYWlsLmNvbSIsImV4cCI6MTc1NTg5NTc4OSwibmFtZSI6IiIsInVpZCI6MX0.X5yO-faRk3vEtOHeSS_dlOSDOiUYWS2EJ2LB7v2CvPw"
	valid, userID, err := service.ValidateJWTToken(example_token)
	if err != nil {
		t.Errorf("Error validating token: %v", err)
	} else {
		t.Logf("Token validation result: %v and %v", valid, userID)
	}
}
