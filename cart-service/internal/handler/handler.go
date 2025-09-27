package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/vsespontanno/eCommerce/cart-service/internal/domain/models"
	"github.com/vsespontanno/eCommerce/cart-service/internal/handler/middleware"
	"go.uber.org/zap"
)

type CartServiceInterface interface {
	Cart(ctx context.Context, userID int64) (*models.Cart, error)
}

type OrderServiceInterface interface {
	GetSelectedProducts(ctx context.Context, userID int64) (map[int64]int64, error)
	SelectProduct(ctx context.Context, userID int64, productID int64) error
	UnselectProduct(ctx context.Context, userID int64, productID int64) error
	ReserveProducts(ctx context.Context, userID int64) error
	ReleaseProducts(ctx context.Context, userID int64) error
	ConfirmOrder(ctx context.Context, userID int64) error
	CancelOrder(ctx context.Context, userID int64) error
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

	router.Handle("/cart/selected",
		h.rateLimiter.RateLimitMiddleware(
			middleware.AuthMiddleware(http.HandlerFunc(h.GetSelected), h.grpcAuthClient),
		),
	).Methods(http.MethodGet)

	router.Handle("/cart/selected/{id}",
		h.rateLimiter.RateLimitMiddleware(
			middleware.AuthMiddleware(http.HandlerFunc(h.SelectProduct), h.grpcAuthClient),
		),
	).Methods(http.MethodPost)

	router.Handle("/cart/selected/{id}",
		h.rateLimiter.RateLimitMiddleware(
			middleware.AuthMiddleware(http.HandlerFunc(h.UnselectProduct), h.grpcAuthClient),
		),
	).Methods(http.MethodDelete)

	// SAGA endpoints
	router.Handle("/cart/checkout",
		h.rateLimiter.RateLimitMiddleware(
			middleware.AuthMiddleware(http.HandlerFunc(h.Checkout), h.grpcAuthClient),
		),
	).Methods(http.MethodPost)

	router.Handle("/cart/confirm",
		h.rateLimiter.RateLimitMiddleware(
			middleware.AuthMiddleware(http.HandlerFunc(h.ConfirmOrder), h.grpcAuthClient),
		),
	).Methods(http.MethodPost)

	router.Handle("/cart/cancel",
		h.rateLimiter.RateLimitMiddleware(
			middleware.AuthMiddleware(http.HandlerFunc(h.CancelOrder), h.grpcAuthClient),
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

// Новые handler методы для SAGA
func (h *Handler) GetSelected(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "Failed to get user ID from context", http.StatusInternalServerError)
		return
	}

	selected, err := h.orderService.GetSelectedProducts(ctx, userID)
	if err != nil {
		h.sugarLogger.Errorf("Failed to get selected products: %v", err)
		http.Error(w, "Failed to get selected products", http.StatusInternalServerError)
		return
	}

	// Преобразуем в формат SelectedCart
	var items []models.SelectedProduct
	for productID, quantity := range selected {
		items = append(items, models.SelectedProduct{
			ProductID: productID,
			Quantity:  quantity,
		})
	}

	selectedCart := models.SelectedCart{
		UserID: userID,
		Items:  items,
		Status: models.StatusDraft,
	}

	writeJSON(w, http.StatusOK, selectedCart)
}

func (h *Handler) SelectProduct(w http.ResponseWriter, r *http.Request) {
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

	// Проверяем, что товар есть в wishlist
	cart, err := h.cartService.Cart(ctx, userID)
	if err != nil {
		http.Error(w, "Failed to get wishlist", http.StatusInternalServerError)
		return
	}

	// Проверяем, что товар есть в wishlist
	found := false
	for _, item := range cart.Items {
		if item.ProductID == int64(intID) {
			found = true
			break
		}
	}
	if !found {
		http.Error(w, "Product not in wishlist", http.StatusBadRequest)
		return
	}

	// Добавляем в selection
	err = h.orderService.SelectProduct(ctx, userID, int64(intID))
	if err != nil {
		if errors.Is(err, models.ErrTooManyProductsOfOneType) {
			writeJSON(w, http.StatusUnprocessableEntity, "You cannot select more than 100 products of one type")
			return
		}
		http.Error(w, "Error while selecting product", http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, "Product selected successfully")
}

func (h *Handler) UnselectProduct(w http.ResponseWriter, r *http.Request) {
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

	err = h.orderService.UnselectProduct(ctx, userID, int64(intID))
	if err != nil {
		http.Error(w, "Error while unselecting product", http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, "Product unselected successfully")
}

func (h *Handler) Checkout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "Failed to get user ID from context", http.StatusInternalServerError)
		return
	}

	// Проверяем, что есть выбранные товары
	selected, err := h.orderService.GetSelectedProducts(ctx, userID)
	if err != nil {
		http.Error(w, "Failed to get selected products", http.StatusInternalServerError)
		return
	}

	if len(selected) == 0 {
		writeJSON(w, http.StatusBadRequest, "No products selected for checkout")
		return
	}

	// Резервируем товары
	err = h.orderService.ReserveProducts(ctx, userID)
	if err != nil {
		http.Error(w, "Failed to reserve products", http.StatusInternalServerError)
		return
	}

	// Генерируем order ID (в реальном проекте это будет UUID)
	orderID := fmt.Sprintf("order_%d_%d", userID, time.Now().Unix())

	response := models.CheckoutResponse{
		OrderID: orderID,
		Status:  models.StatusReserved,
		Message: "Products reserved successfully",
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) ConfirmOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "Failed to get user ID from context", http.StatusInternalServerError)
		return
	}

	err := h.orderService.ConfirmOrder(ctx, userID)
	if err != nil {
		http.Error(w, "Failed to confirm order", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, "Order confirmed successfully")
}

func (h *Handler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "Failed to get user ID from context", http.StatusInternalServerError)
		return
	}

	// Освобождаем резерв
	err := h.orderService.ReleaseProducts(ctx, userID)
	if err != nil {
		h.sugarLogger.Errorf("Failed to release products: %v", err)
	}

	// Отменяем заказ
	err = h.orderService.CancelOrder(ctx, userID)
	if err != nil {
		http.Error(w, "Failed to cancel order", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, "Order cancelled successfully")
}

func writeJSON(rw http.ResponseWriter, status int, v any) error {
	rw.WriteHeader(status)
	rw.Header().Add("Content-Type", "application/json")
	return json.NewEncoder(rw).Encode(v)
}
