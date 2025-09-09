package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vsespontanno/eCommerce/cart-service/internal/client"
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
// POST /cart/{id} - adding item into the order; quantity = all (redis)
// POST /cart/{id}/quantity - adding n-quantity of item (redis)
// POST(maybe) /cart/order - gRPC req to SAGA orchestrator microservice "Order" who will make transaction
func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.Handle("/cart", middleware.AuthMiddleware(http.HandlerFunc(h.GetCart), h.grpcAuthClient)).Methods(http.MethodGet)
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
		if err == sql.ErrNoRows {
			writeJSON(w, http.StatusOK, "Your cart is empty")
			return
		}
		h.sugarLogger.Errorf("Failed to get cart: %v", err)
		http.Error(w, "Failed to get cart", http.StatusInternalServerError)
		return
	}
	productIDs := []int64{}
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
	w.Write(serialized)
}

func writeJSON(rw http.ResponseWriter, status int, v any) error {
	rw.WriteHeader(status)
	rw.Header().Add("Content-Type", "application/json")
	return json.NewEncoder(rw).Encode(v)
}
