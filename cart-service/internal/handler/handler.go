package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/vsespontanno/eCommerce/cart-service/internal/domain/models"
	"github.com/vsespontanno/eCommerce/cart-service/internal/handler/middleware"
	"go.uber.org/zap"
)

type CartServiceInterface interface {
	Cart(ctx context.Context, userID int64) (*models.Cart, error)
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
	ValidateToken(ctx context.Context, token string) (*models.TokenResponse, error)
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

}

func (h *Handler) GetCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "failed to get user ID from context", http.StatusInternalServerError)
		return
	}
	cart, err := h.cartService.Cart(ctx, userID)
	if err != nil {
		// если это пустая корзина
		if errors.Is(err, models.ErrNoCartFound) {
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"message": "Your cart is empty",
				"items":   []models.CartItem{},
			})
			return
		}

		h.sugarLogger.Errorf("failed to get cart: %v", err)
		http.Error(w, "failed to get cart", http.StatusInternalServerError)
		return
	}

	if len(cart.Items) == 0 {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"message": "Your cart is empty",
			"items":   []models.CartItem{},
		})
		return
	}

	writeJSON(w, http.StatusOK, cart)
}

func (h *Handler) ClearCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "Failed to get user ID from context", http.StatusInternalServerError)
		return
	}
	err := h.cartService.ClearCart(ctx, userID)
	if err != nil {
		http.Error(w, "Error while clearing cart", http.StatusBadRequest)
		return
	}
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
		return
	}
	vars := mux.Vars(r)
	stringID := vars["id"]
	intID, err := strconv.Atoi(stringID)
	if err != nil || intID <= 0 || intID > 1000000 {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}
	err = h.cartService.DeleteProductFromCart(ctx, userID, int64(intID))
	if err != nil {
		http.Error(w, "Error while removing product", http.StatusBadRequest)
		return
	}
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
		return
	}
	vars := mux.Vars(r)
	stringID := vars["id"]
	intID, err := strconv.Atoi(stringID)
	if err != nil || intID <= 0 || intID > 1000000 {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}
	err = h.cartService.AddProductToCart(ctx, userID, int64(intID))
	if err != nil {
		if errors.Is(err, models.ErrTooManyProductsOfOneType) {
			writeJSON(w, http.StatusUnprocessableEntity, "You cannot add more than 100 products of one")
			return
		}
		http.Error(w, "Error while adding product", http.StatusBadRequest)
		return
	}
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
		return
	}
	vars := mux.Vars(r)
	stringID := vars["id"]
	intID, err := strconv.Atoi(stringID)
	if err != nil || intID <= 0 || intID > 1000000 {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}
	err = h.cartService.Decrement(ctx, userID, int64(intID))
	if err != nil {
		http.Error(w, "Error while removing product", http.StatusBadRequest)
		return
	}
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
		return
	}
	orderID, err := h.checkouter.Checkout(ctx, userID)
	if err != nil {
		http.Error(w, "Error while checking out", http.StatusBadRequest)
		return
	}
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

func writeJSON(rw http.ResponseWriter, status int, v any) error {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(status)
	return json.NewEncoder(rw).Encode(v)
}
