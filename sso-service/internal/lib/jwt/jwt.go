package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vsespontanno/eCommerce/sso-service/internal/domain/models"
)

func NewToken(user models.User, jwtSecret string, duration time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = user.ID
	claims["email"] = user.Email
	claims["exp"] = time.Now().Add(duration).Unix()
	claims["name"] = user.FirstName
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
