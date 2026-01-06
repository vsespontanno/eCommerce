package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/vsespontanno/eCommerce/services/products-service/internal/infrastructure/metrics"
)

// MetricsMiddleware собирает метрики для HTTP запросов
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		metrics.ActiveConnections.Inc()
		defer metrics.ActiveConnections.Dec()

		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(wrapped.statusCode)

		metrics.HTTPRequestsTotal.WithLabelValues(
			r.Method,
			r.URL.Path,
			status,
		).Inc()

		metrics.HTTPRequestDuration.WithLabelValues(
			r.Method,
			r.URL.Path,
		).Observe(duration)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
