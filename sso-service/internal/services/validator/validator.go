package validator

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
)

type Validator struct {
	jwtSecret string
}

func New(jwtSecret string) *Validator {
	return &Validator{
		jwtSecret: jwtSecret,
	}
}

func (v *Validator) ValidateJWTToken(token string) (valid bool, userID float64, err error) {
	if token == "" {
		return false, 0, ErrInvalidToken
	}

	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(v.jwtSecret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return false, 0, ErrTokenExpired
		}
		return false, 0, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	if !parsedToken.Valid {
		return false, 0, ErrInvalidToken
	}

	// Safely extract claims
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return false, 0, ErrInvalidToken
	}

	// Safely extract user ID
	uidRaw, ok := claims["uid"]
	if !ok {
		return false, 0, ErrInvalidToken
	}

	uid, ok := uidRaw.(float64)
	if !ok {
		return false, 0, ErrInvalidToken
	}

	return true, uid, nil
}
