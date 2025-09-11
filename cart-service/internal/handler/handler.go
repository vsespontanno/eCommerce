package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/vsespontanno/eCommerce/cart-service/internal/client"
	"github.com/vsespontanno/eCommerce/cart-service/internal/domain/models"
	"github.com/vsespontanno/eCommerce/cart-service/internal/handler/middleware"
	"github.com/vsespontanno/eCommerce/cart-service/internal/service"
	"go.uber.org/zap"
)

type Handler struct {
	cartService    *service.CartService
	orderService   *service.OrderService
	sugarLogger    *zap.SugaredLogger
	grpcAuthClient *client.JwtClient
}

func New(cartService *service.CartService, sugarLogger *zap.SugaredLogger, grpcAuthClient *client.JwtClient, orderService *service.OrderService) *Handler {
	return &Handler{
		cartService:    cartService,
		sugarLogger:    sugarLogger,
		grpcAuthClient: grpcAuthClient,
		orderService:   orderService,
	}
}

// GET /cart - every item from cart that user can add into his order (postgresql)
// GET /cart/{id} - getting product from cart (without it for some time)
// UPD: GRPC REQ TO PRODUCT SERVICE FOR BOTH
// POST /cart/{id} - adding item into the order; quantity = not (__all__), only one item. if add again = +1 till 10 (redis)
// POST(maybe) /cart/order - gRPC req to SAGA orchestrator microservice "Order" who will make transaction
func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.Handle("/cart", middleware.AuthMiddleware(http.HandlerFunc(h.GetCart), h.grpcAuthClient)).Methods(http.MethodGet)
	router.Handle("/cart/{id}", middleware.AuthMiddleware(http.HandlerFunc(h.AddProduct), h.grpcAuthClient)).Methods(http.MethodPost)
}

func (h *Handler) GetCart(w http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "Failed to get user ID from context", http.StatusInternalServerError)
		return
	}
	cart, err := h.cartService.Cart(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusOK, "Your cart is empty")
			return
		}
		h.sugarLogger.Errorf("Failed to get cart: %v", err)
		http.Error(w, "Failed to get cart", http.StatusInternalServerError)
		return
	}
	var productIDs []int64
	for _, item := range cart.Items {
		productIDs = append(productIDs, item.ProductID)
	}

	err = h.orderService.AddAllProducts(ctx, userID, productIDs)
	if err != nil {
		h.sugarLogger.Errorf("Failed to add products to order: %v", err)
		http.Error(w, "Failed to add products to order", http.StatusInternalServerError)
		return
	}

	serialized, err := json.Marshal(cart)
	if err != nil {
		h.sugarLogger.Errorf("Failed to serialize product: %v", err)
		http.Error(w, "Failed to serialize product", http.StatusInternalServerError)
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
	ctx := context.TODO()
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "Failed to get user ID from context", http.StatusInternalServerError)
		return
	}
	vars := mux.Vars(r)
	stringID := vars["id"]
	intID, err := strconv.Atoi(stringID)
	if err != nil {
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

func writeJSON(rw http.ResponseWriter, status int, v any) error {
	rw.WriteHeader(status)
	rw.Header().Add("Content-Type", "application/json")
	return json.NewEncoder(rw).Encode(v)
}
