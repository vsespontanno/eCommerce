package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vsespontanno/eCommerce/user-service/internal/models"
)

type JwtService struct {
	Secret string
}

var ErrInvalidToken = errors.New("invalid token")

func NewJwtService(secret string) (*JwtService, error) {
	return &JwtService{Secret: secret}, nil
}

func (s *JwtService) GenerateToken(user *models.User, duration time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = user.ID
	claims["email"] = user.Email
	claims["exp"] = time.Now().Add(duration).Unix()

	tokenString, err := token.SignedString([]byte(s.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *JwtService) ValidateToken(tokenString string) (bool, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.Secret), nil
	})
	if err != nil {
		return false, err
	}

	if !token.Valid {
		return false, ErrInvalidToken
	}

	_, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false, ErrInvalidToken
	}

	return true, nil
}
