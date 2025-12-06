package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/ulule/limiter/v3"
	redisstore "github.com/ulule/limiter/v3/drivers/store/redis"
)

type RateLimiter struct {
	limiter *limiter.Limiter
}

func NewRateLimiter(rdb *redis.Client, rateLimitRPS int) *RateLimiter {
	// Создаем Redis store для rate limiter
	store, err := redisstore.NewStoreWithOptions(rdb, limiter.StoreOptions{
		Prefix:   "rate_limit",
		MaxRetry: 3,
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to create Redis store for rate limiter: %v", err))
	}

	// Создаем rate limiter с настройками
	rate := limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  int64(rateLimitRPS),
	}

	instance := limiter.New(store, rate)

	return &RateLimiter{
		limiter: instance,
	}
}

func (rl *RateLimiter) RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Rate limiter работает ДО auth middleware, поэтому используем IP или токен
		var key string

		// Пытаемся получить userID из контекста (если auth уже отработал)
		userID, ok := r.Context().Value(UserIDKey).(int64)
		if ok && userID != 0 {
			key = fmt.Sprintf("user_%d", userID)
		} else {
			// Если auth еще не отработал, используем IP адрес
			key = fmt.Sprintf("ip_%s", r.RemoteAddr)
		}

		// Проверяем лимит
		context, err := rl.limiter.Get(r.Context(), key)
		if err != nil {
			// В случае ошибки разрешаем запрос (fail open)
			next.ServeHTTP(w, r)
			return
		}

		// Проверяем, превышен ли лимит
		if context.Reached {
			w.Header().Set("X-RateLimit-Limit", strconv.FormatInt(context.Limit, 10))
			w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(context.Remaining, 10))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(context.Reset, 10))

			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		// Добавляем заголовки с информацией о лимитах
		w.Header().Set("X-RateLimit-Limit", strconv.FormatInt(context.Limit, 10))
		w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(context.Remaining, 10))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(context.Reset, 10))

		next.ServeHTTP(w, r)
	})
}
