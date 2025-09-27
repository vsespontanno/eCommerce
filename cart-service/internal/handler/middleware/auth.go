package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/vsespontanno/eCommerce/cart-service/internal/domain/models"
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
	ValidateToken(ctx context.Context, token string) (*models.TokenResponse, error)
}

// AuthMiddleware проверяет JWT токен и добавляет информацию о пользователе в контекст
func AuthMiddleware(next http.HandlerFunc, grpcClient ValidatorInterface) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Получаем токен из заголовка
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, ErrNoAuthHeader.Error(), http.StatusUnauthorized)
			return
		}
		// Проверяем формат заголовка
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, ErrInvalidAuthHeader.Error(), http.StatusUnauthorized)
			return
		}
		// Проверяем токен
		valid, err := grpcClient.ValidateToken(r.Context(), parts[1])
		if err != nil || !valid.Valid {
			http.Error(w, ErrInvalidToken.Error(), http.StatusUnauthorized)
			return
		}
		// Добавляем информацию о пользователе в контекст
		ctx := context.WithValue(r.Context(), UserIDKey, valid.UserID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
