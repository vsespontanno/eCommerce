package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/vsespontanno/eCommerce/services/cart-service/internal/infrastructure/client/grpc/jwt/dto"
)

var (
	ErrNoAuthHeader      = errors.New("no authorization header")
	ErrInvalidAuthHeader = errors.New("invalid authorization header")
	ErrInvalidToken      = errors.New("invalid token")
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
)

type ValidatorInterface interface {
	ValidateToken(ctx context.Context, token string) (*dto.TokenResponse, error)
}

func AuthMiddleware(next http.HandlerFunc, grpcClient ValidatorInterface) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, ErrNoAuthHeader.Error(), http.StatusUnauthorized)
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, ErrInvalidAuthHeader.Error(), http.StatusUnauthorized)
			return
		}
		valid, err := grpcClient.ValidateToken(r.Context(), parts[1])
		if err != nil || !valid.Valid {
			http.Error(w, ErrInvalidToken.Error(), http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), UserIDKey, valid.UserID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
