package handler

import (
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type Handler struct {
	sugarLogger *zap.SugaredLogger
}

// GET /cart - every item from cart that user can add into his order (postgresql)
// GET /cart/{id} - getting product from cart (postgresql)
// POST /cart/{id} - adding item into the order; quantity = all (redis)
// POST /cart/{id}/quantity - adding n-quantity of item (redis)
// POST(maybe) /cart/order - gRPC req to SAGA orchestrator microservice "Order" who will make transaction
func (h *Handler) RegisterRoutes(router *mux.Router) {
}
