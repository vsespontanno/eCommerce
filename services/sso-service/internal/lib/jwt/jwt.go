package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vsespontanno/eCommerce/services/sso-service/internal/domain/models"
)

type TokenPair struct {
	Token     string
	ExpiresAt int64
}

func NewToken(user models.User, jwtSecret string, duration time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	expiresAt := time.Now().Add(duration).Unix()

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("failed to get claims")
	}
	claims["uid"] = user.ID
	claims["email"] = user.Email
	claims["exp"] = expiresAt
	claims["iat"] = time.Now().Unix()

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func NewTokenWithExpiry(user models.User, jwtSecret string, duration time.Duration) (TokenPair, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	expiresAt := time.Now().Add(duration).Unix()

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return TokenPair{}, fmt.Errorf("failed to get claims")
	}
	claims["uid"] = user.ID
	claims["email"] = user.Email
	claims["exp"] = expiresAt
	claims["iat"] = time.Now().Unix()

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		Token:     tokenString,
		ExpiresAt: expiresAt,
	}, nil
}
