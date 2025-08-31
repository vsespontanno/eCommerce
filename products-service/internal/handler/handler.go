package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/vsespontanno/eCommerce/products-service/internal/client"
	"github.com/vsespontanno/eCommerce/products-service/internal/domain/models"
	"github.com/vsespontanno/eCommerce/products-service/internal/handler/middleware"
	"github.com/vsespontanno/eCommerce/products-service/internal/repository/postgres"
	"go.uber.org/zap"
)

type Handler struct {
	cartStore    *postgres.CartStore
	productStore *postgres.ProductStore
	sugarLogger  *zap.SugaredLogger
	grpcClient   *client.JwtClient
}

func New(cartStore *postgres.CartStore, productStore *postgres.ProductStore, sugarLogger *zap.SugaredLogger, grpcClient *client.JwtClient) *Handler {
	return &Handler{
		cartStore:    cartStore,
		productStore: productStore,
		sugarLogger:  sugarLogger,
		grpcClient:   grpcClient,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	//it would be in potential , but know for beginning we create simple thing
	router.HandleFunc("/product/{id}", h.GetProduct).Methods(http.MethodGet)
	router.HandleFunc("/product", h.GetProducts).Methods(http.MethodGet)
	router.Handle("/product/{id}/add-to-cart", middleware.AuthMiddleware(http.HandlerFunc(h.AddProductToCart), h.grpcClient)).Methods(http.MethodPost)
}

// TODO: pagination
func (h *Handler) GetProducts(w http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()
	products, err := h.productStore.GetProducts(ctx)
	if err != nil {
		h.sugarLogger.Errorf("Failed to get products: %v", err)
		http.Error(w, "Failed to get products", http.StatusInternalServerError)
		return
	}

	h.sugarLogger.Infof("Retrieved products: %v", products)
	serialized, err := json.Marshal(products)
	if err != nil {
		h.sugarLogger.Errorf("Failed to serialize products: %v", err)
		http.Error(w, "Failed to serialize products", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(serialized)
}

func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GetProduct called")
	ctx := context.TODO()
	vars := mux.Vars(r)
	string_id := vars["id"]
	int_id, err := strconv.Atoi(string_id)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}
	product, err := h.productStore.GetProductByID(ctx, int64(int_id))
	if err != nil {
		if err == models.ErrNoProductFound {
			http.Error(w, "Product not found", http.StatusNotFound)
		} else {
			h.sugarLogger.Errorf("Failed to get product: %v", err)
			http.Error(w, "Failed to get product", http.StatusInternalServerError)
		}
		return
	}

	h.sugarLogger.Infof("Retrieved product: %v", product)

	serialized, err := json.Marshal(product)
	if err != nil {
		h.sugarLogger.Errorf("Failed to serialize product: %v", err)
		http.Error(w, "Failed to serialize product", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(serialized)
}

func (h *Handler) AddProductToCart(w http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()
	vars := mux.Vars(r)
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "Failed to get user ID from context", http.StatusInternalServerError)
		return
	}
	string_id := vars["id"]
	int_id, err := strconv.Atoi(string_id)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}
	product, err := h.productStore.GetProductByID(ctx, int64(int_id))
	if err != nil {
		if err == models.ErrNoProductFound {
			http.Error(w, "Product not found", http.StatusNotFound)
		} else {
			h.sugarLogger.Errorf("Failed to get product: %v", err)
			http.Error(w, "Failed to get product", http.StatusInternalServerError)
		}
		return
	}

	h.sugarLogger.Infof("Retrieved product: %v", product)
	_, err = h.cartStore.UpsertProductToCart(ctx, userID, product.ID)
	if err != nil {
		h.sugarLogger.Errorf("Failed to upsert product to cart: %v", err)
		http.Error(w, "Failed to upsert product to cart", http.StatusInternalServerError)
		return
	}
	serialized, err := json.Marshal(product)
	if err != nil {
		h.sugarLogger.Errorf("Failed to serialize product: %v", err)
		http.Error(w, "Failed to serialize product", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(serialized)
}
