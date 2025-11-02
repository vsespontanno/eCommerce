package handler

import (
	"context"
	"database/sql"
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
}

type OrderServiceInterface interface {
	AddProductToCart(ctx context.Context, userID int64, productID int64) error
}

type RateLimiterInterface interface {
	RateLimitMiddleware(next http.Handler) http.Handler
}

type ValidatorInterface interface {
	ValidateToken(ctx context.Context, token string) (*models.TokenResponse, error)
}

type Handler struct {
	cartService    CartServiceInterface
	orderService   OrderServiceInterface
	sugarLogger    *zap.SugaredLogger
	grpcAuthClient ValidatorInterface
	rateLimiter    RateLimiterInterface
}

func New(cartService CartServiceInterface, sugarLogger *zap.SugaredLogger, grpcAuthClient ValidatorInterface, orderService OrderServiceInterface, rateLimiter RateLimiterInterface) *Handler {
	return &Handler{
		cartService:    cartService,
		sugarLogger:    sugarLogger,
		grpcAuthClient: grpcAuthClient,
		orderService:   orderService,
		rateLimiter:    rateLimiter,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.Handle("/cart",
		h.rateLimiter.RateLimitMiddleware(
			middleware.AuthMiddleware(http.HandlerFunc(h.GetCart), h.grpcAuthClient),
		),
	).Methods(http.MethodGet)

	router.Handle("/cart/{id}",
		h.rateLimiter.RateLimitMiddleware(
			middleware.AuthMiddleware(http.HandlerFunc(h.AddProduct), h.grpcAuthClient),
		),
	).Methods(http.MethodPost)

	// SAGA endpoints
	router.Handle("/cart/checkout",
		h.rateLimiter.RateLimitMiddleware(
			middleware.AuthMiddleware(http.HandlerFunc(h.Checkout), h.grpcAuthClient),
		),
	).Methods(http.MethodPost)

}

func (h *Handler) GetCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context() // Используем контекст из запроса
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "Failed to get user ID from context", http.StatusInternalServerError)
		return
	}

	// Получаем только wishlist (PostgreSQL) - НЕ добавляем в Redis!
	cart, err := h.cartService.Cart(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusOK, "Your wishlist is empty")
			return
		}
		h.sugarLogger.Errorf("Failed to get cart: %v", err)
		http.Error(w, "Failed to get cart", http.StatusInternalServerError)
		return
	}

	serialized, err := json.Marshal(cart)
	if err != nil {
		h.sugarLogger.Errorf("Failed to serialize cart: %v", err)
		http.Error(w, "Failed to serialize cart", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(serialized)
	if err != nil {
		h.sugarLogger.Errorf("Failed to write response: %v", err)
		return
	}
}

func (h *Handler) AddProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context() // Используем контекст из запроса
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
	err = h.orderService.AddProductToCart(ctx, userID, int64(intID))
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

func (h *Handler) Checkout(w http.ResponseWriter, r *http.Request) {
	//TODO: request to saga orch service
}

func writeJSON(rw http.ResponseWriter, status int, v any) error {
	rw.WriteHeader(status)
	rw.Header().Add("Content-Type", "application/json")
	return json.NewEncoder(rw).Encode(v)
}
