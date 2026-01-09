package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP метрики
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cart_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cart_http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
		},
		[]string{"method", "endpoint"},
	)

	ActiveConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "cart_active_connections",
			Help: "Number of active connections",
		},
	)

	// Бизнес-метрики корзины
	CartOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cart_operations_total",
			Help: "Total number of cart operations",
		},
		[]string{"operation", "status"},
	)

	CartItemsCount = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cart_items_count",
			Help:    "Number of items in cart",
			Buckets: []float64{0, 1, 2, 5, 10, 20, 50, 100},
		},
		[]string{"user_id"},
	)

	CartTotalValue = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cart_total_value",
			Help:    "Total value of cart in currency units",
			Buckets: []float64{0, 100, 500, 1000, 5000, 10000, 50000, 100000},
		},
		[]string{"user_id"},
	)

	CheckoutTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cart_checkout_total",
			Help: "Total number of checkout operations",
		},
		[]string{"status"},
	)

	ProductAddedToCartTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cart_product_added_total",
			Help: "Total number of products added to cart",
		},
		[]string{"product_id"},
	)

	ProductRemovedFromCartTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cart_product_removed_total",
			Help: "Total number of products removed from cart",
		},
		[]string{"product_id"},
	)

	// Redis метрики
	RedisOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cart_redis_operations_total",
			Help: "Total number of Redis operations",
		},
		[]string{"operation", "status"},
	)

	RedisOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cart_redis_operation_duration_seconds",
			Help:    "Redis operation latency in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"operation"},
	)

	// Database метрики
	DBQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cart_db_query_duration_seconds",
			Help:    "Database query latency in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"operation", "table"},
	)

	DBQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cart_db_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table", "status"},
	)

	// Rate Limiter метрики
	RateLimitHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cart_rate_limit_hits_total",
			Help: "Total number of rate limit hits",
		},
		[]string{"endpoint", "result"},
	)
)
