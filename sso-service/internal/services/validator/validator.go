package validator

import "github.com/golang-jwt/jwt/v5"

type Validator struct {
	jwtSecret string
}

func New(jwtSecret string) *Validator {
	return &Validator{
		jwtSecret: jwtSecret,
	}
}

func (v *Validator) ValidateJWTToken(token string) (valid bool, userID float64, Werr error) {
	if token == "" {
		return false, 0.0, nil
	}
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(v.jwtSecret), nil
	})

	if err != nil {
		return false, 0.0, err
	}
	return parsedToken.Valid, parsedToken.Claims.(jwt.MapClaims)["uid"].(float64), nil
}
