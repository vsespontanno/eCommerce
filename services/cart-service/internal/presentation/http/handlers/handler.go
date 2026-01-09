package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/vsespontanno/eCommerce/services/cart-service/internal/domain/apperrors"
	"github.com/vsespontanno/eCommerce/services/cart-service/internal/domain/cart/entity"
	"github.com/vsespontanno/eCommerce/services/cart-service/internal/infrastructure/client/grpc/jwt/dto"
	"github.com/vsespontanno/eCommerce/services/cart-service/internal/infrastructure/metrics"
	"github.com/vsespontanno/eCommerce/services/cart-service/internal/presentation/http/handlers/middleware"
	"go.uber.org/zap"
)

type CartServiceInterface interface {
	Cart(ctx context.Context, userID int64) (*entity.Cart, error)
	AddProductToCart(ctx context.Context, userID int64, productID int64) error
	Increment(ctx context.Context, userID int64, productID int64) error
	Decrement(ctx context.Context, userID int64, productID int64) error
	ClearCart(ctx context.Context, userID int64) error
	DeleteProductFromCart(ctx context.Context, userID int64, productID int64) error
}

type RateLimiterInterface interface {
	RateLimitMiddleware(next http.Handler) http.Handler
}

type ValidatorInterface interface {
	ValidateToken(ctx context.Context, token string) (*dto.TokenResponse, error)
}

type Checkouter interface {
	Checkout(ctx context.Context, userID int64) (string, error)
}

type Handler struct {
	cartService    CartServiceInterface
	sugarLogger    *zap.SugaredLogger
	grpcAuthClient ValidatorInterface
	rateLimiter    RateLimiterInterface
	checkouter     Checkouter
}

func New(cartService CartServiceInterface, sugarLogger *zap.SugaredLogger,
	grpcAuthClient ValidatorInterface,
	rateLimiter RateLimiterInterface, checkouter Checkouter) *Handler {
	return &Handler{
		cartService:    cartService,
		sugarLogger:    sugarLogger,
		grpcAuthClient: grpcAuthClient,
		rateLimiter:    rateLimiter,
		checkouter:     checkouter,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	// Подключаем metrics middleware ко всем роутам
	router.Use(middleware.MetricsMiddleware)

	router.Handle("/cart",
		h.rateLimiter.RateLimitMiddleware(
			middleware.AuthMiddleware(http.HandlerFunc(h.GetCart), h.grpcAuthClient),
		),
	).Methods(http.MethodGet)

	router.Handle("/cart/{id}/increment",
		h.rateLimiter.RateLimitMiddleware(
			middleware.AuthMiddleware(http.HandlerFunc(h.IncrementProduct), h.grpcAuthClient),
		),
	).Methods(http.MethodPatch)

	router.Handle("/cart/{id}/decrement",
		h.rateLimiter.RateLimitMiddleware(
			middleware.AuthMiddleware(http.HandlerFunc(h.DecrementProduct), h.grpcAuthClient),
		),
	).Methods(http.MethodPatch)

	router.Handle("/cart/{id}",
		h.rateLimiter.RateLimitMiddleware(
			middleware.AuthMiddleware(http.HandlerFunc(h.RemoveProduct), h.grpcAuthClient),
		),
	).Methods(http.MethodDelete)

	router.Handle("/cart",
		h.rateLimiter.RateLimitMiddleware(
			middleware.AuthMiddleware(http.HandlerFunc(h.ClearCart), h.grpcAuthClient),
		),
	).Methods(http.MethodDelete)

	// SAGA endpoints
	router.Handle("/cart/order/checkout",
		h.rateLimiter.RateLimitMiddleware(
			middleware.AuthMiddleware(http.HandlerFunc(h.Checkout), h.grpcAuthClient),
		),
	).Methods(http.MethodPost)

	router.HandleFunc("/health", h.HealthCheck).Methods(http.MethodGet)

	// Prometheus metrics endpoint
	router.Handle("/metrics", promhttp.Handler())
}

func (h *Handler) GetCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "failed to get user ID from context", http.StatusInternalServerError)
		metrics.CartOperationsTotal.WithLabelValues("get_cart", "error").Inc()
		return
	}
	cart, err := h.cartService.Cart(ctx, userID)
	if err != nil {
		// если это пустая корзина
		if errors.Is(err, apperrors.ErrNoCartFound) {
			metrics.CartOperationsTotal.WithLabelValues("get_cart", "empty").Inc()
			if writeErr := writeJSON(w, http.StatusOK, map[string]interface{}{
				"message": "Your cart is empty",
				"items":   []entity.CartItem{},
			}); writeErr != nil {
				h.sugarLogger.Errorw("failed to write empty cart response", "error", writeErr)
			}
			return
		}

		h.sugarLogger.Errorf("failed to get cart: %v", err)
		metrics.CartOperationsTotal.WithLabelValues("get_cart", "error").Inc()
		http.Error(w, "failed to get cart", http.StatusInternalServerError)
		return
	}

	// Записываем метрики корзины
	metrics.CartOperationsTotal.WithLabelValues("get_cart", "success").Inc()
	metrics.CartItemsCount.WithLabelValues(strconv.FormatInt(userID, 10)).Observe(float64(len(cart.Items)))

	// Подсчитываем общую стоимость корзины
	var totalValue int64
	for _, item := range cart.Items {
		totalValue += item.Price * int64(item.Quantity)
	}
	metrics.CartTotalValue.WithLabelValues(strconv.FormatInt(userID, 10)).Observe(float64(totalValue))

	if len(cart.Items) == 0 {
		if writeErr := writeJSON(w, http.StatusOK, map[string]interface{}{
			"message": "Your cart is empty",
			"items":   []entity.CartItem{},
		}); writeErr != nil {
			h.sugarLogger.Errorw("failed to write empty cart response", "error", writeErr)
		}
		return
	}

	if writeErr := writeJSON(w, http.StatusOK, cart); writeErr != nil {
		h.sugarLogger.Errorw("failed to write cart response", "error", writeErr)
	}
}

func (h *Handler) ClearCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "Failed to get user ID from context", http.StatusInternalServerError)
		metrics.CartOperationsTotal.WithLabelValues("clear_cart", "error").Inc()
		return
	}
	err := h.cartService.ClearCart(ctx, userID)
	if err != nil {
		http.Error(w, "Error while clearing cart", http.StatusBadRequest)
		metrics.CartOperationsTotal.WithLabelValues("clear_cart", "error").Inc()
		return
	}
	metrics.CartOperationsTotal.WithLabelValues("clear_cart", "success").Inc()
	err = writeJSON(w, http.StatusOK, "")
	if err != nil {
		h.sugarLogger.Errorf("Failed to write response: %v", err)
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}
func (h *Handler) RemoveProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "Failed to get user ID from context", http.StatusInternalServerError)
		metrics.CartOperationsTotal.WithLabelValues("remove_product", "error").Inc()
		return
	}
	vars := mux.Vars(r)
	stringID := vars["id"]
	intID, err := strconv.Atoi(stringID)
	if err != nil || intID <= 0 || intID > 1000000 {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		metrics.CartOperationsTotal.WithLabelValues("remove_product", "invalid_id").Inc()
		return
	}
	err = h.cartService.DeleteProductFromCart(ctx, userID, int64(intID))
	if err != nil {
		http.Error(w, "Error while removing product", http.StatusBadRequest)
		metrics.CartOperationsTotal.WithLabelValues("remove_product", "error").Inc()
		return
	}

	// Записываем метрику удаления товара
	metrics.ProductRemovedFromCartTotal.WithLabelValues(stringID).Inc()
	metrics.CartOperationsTotal.WithLabelValues("remove_product", "success").Inc()

	err = writeJSON(w, http.StatusOK, "")
	if err != nil {
		h.sugarLogger.Errorf("Failed to write response: %v", err)
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) IncrementProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "Failed to get user ID from context", http.StatusInternalServerError)
		metrics.CartOperationsTotal.WithLabelValues("increment_product", "error").Inc()
		return
	}
	vars := mux.Vars(r)
	stringID := vars["id"]
	intID, err := strconv.Atoi(stringID)
	if err != nil || intID <= 0 || intID > 1000000 {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		metrics.CartOperationsTotal.WithLabelValues("increment_product", "invalid_id").Inc()
		return
	}
	err = h.cartService.AddProductToCart(ctx, userID, int64(intID))
	if err != nil {
		if errors.Is(err, apperrors.ErrTooManyProductsOfOneType) {
			metrics.CartOperationsTotal.WithLabelValues("increment_product", "limit_exceeded").Inc()
			if writeErr := writeJSON(w, http.StatusUnprocessableEntity, "You cannot add more than 100 products of one"); writeErr != nil {
				h.sugarLogger.Errorw("failed to write error response", "error", writeErr)
			}
			return
		}
		http.Error(w, "Error while adding product", http.StatusBadRequest)
		metrics.CartOperationsTotal.WithLabelValues("increment_product", "error").Inc()
		return
	}

	// Записываем метрику добавления товара
	metrics.ProductAddedToCartTotal.WithLabelValues(stringID).Inc()
	metrics.CartOperationsTotal.WithLabelValues("increment_product", "success").Inc()

	err = writeJSON(w, http.StatusOK, "")
	if err != nil {
		h.sugarLogger.Errorf("Failed to add product: %v", err)
		http.Error(w, "Failed to add product", http.StatusInternalServerError)
		return
	}

}

func (h *Handler) DecrementProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "Failed to get user ID from context", http.StatusInternalServerError)
		metrics.CartOperationsTotal.WithLabelValues("decrement_product", "error").Inc()
		return
	}
	vars := mux.Vars(r)
	stringID := vars["id"]
	intID, err := strconv.Atoi(stringID)
	if err != nil || intID <= 0 || intID > 1000000 {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		metrics.CartOperationsTotal.WithLabelValues("decrement_product", "invalid_id").Inc()
		return
	}
	err = h.cartService.Decrement(ctx, userID, int64(intID))
	if err != nil {
		http.Error(w, "Error while removing product", http.StatusBadRequest)
		metrics.CartOperationsTotal.WithLabelValues("decrement_product", "error").Inc()
		return
	}

	metrics.CartOperationsTotal.WithLabelValues("decrement_product", "success").Inc()

	err = writeJSON(w, http.StatusOK, "")
	if err != nil {
		h.sugarLogger.Errorf("Failed to remove product: %v", err)
		http.Error(w, "Failed to remove product", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Checkout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "Failed to get user ID from context", http.StatusInternalServerError)
		metrics.CheckoutTotal.WithLabelValues("error").Inc()
		return
	}
	orderID, err := h.checkouter.Checkout(ctx, userID)
	if err != nil {
		http.Error(w, "Error while checking out", http.StatusBadRequest)
		metrics.CheckoutTotal.WithLabelValues("error").Inc()
		return
	}

	// Записываем успешный checkout
	metrics.CheckoutTotal.WithLabelValues("success").Inc()

	err = writeJSON(w, http.StatusAccepted, map[string]interface{}{
		"message": "Order accepted and is being processed",
		"orderId": orderID,
	})
	if err != nil {
		h.sugarLogger.Errorf("Failed to checkout: %v", err)
		http.Error(w, "Failed to checkout", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if err := writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"service": "cart-service",
	}); err != nil {
		h.sugarLogger.Errorw("failed to write health check response", "error", err)
	}
}

func writeJSON(rw http.ResponseWriter, status int, v any) error {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(status)
	return json.NewEncoder(rw).Encode(v)
}
