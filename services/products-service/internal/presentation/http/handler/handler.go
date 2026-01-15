package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/vsespontanno/eCommerce/services/products-service/internal/domain/apperrors"
	"github.com/vsespontanno/eCommerce/services/products-service/internal/domain/products/entity"
	client "github.com/vsespontanno/eCommerce/services/products-service/internal/infrastructure/client/grpc"
	"github.com/vsespontanno/eCommerce/services/products-service/internal/infrastructure/metrics"
	"github.com/vsespontanno/eCommerce/services/products-service/internal/presentation/http/handler/middleware"
	metricsMiddleware "github.com/vsespontanno/eCommerce/services/products-service/internal/presentation/http/middleware"
	"go.uber.org/zap"
)

type CartStorer interface {
	UpsertProductToCart(ctx context.Context, userID int64, productID int64, amountForProduct int64) (int, error)
}

type ProductStorer interface {
	SaveProduct(ctx context.Context, product *entity.Product) error
	GetProducts(ctx context.Context) ([]*entity.Product, error)
	GetProductByID(ctx context.Context, id int64) (*entity.Product, error)
	GetProductsByID(ctx context.Context, ids []int64) ([]*entity.Product, error)
}

type Handler struct {
	cartStore    CartStorer
	productStore ProductStorer
	sugarLogger  *zap.SugaredLogger
	grpcClient   *client.JwtClient
}

func New(cartStore CartStorer, productStore ProductStorer, sugarLogger *zap.SugaredLogger, grpcClient *client.JwtClient) *Handler {
	return &Handler{
		cartStore:    cartStore,
		productStore: productStore,
		sugarLogger:  sugarLogger,
		grpcClient:   grpcClient,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	// Подключаем metrics middleware ко всем роутам
	router.Use(metricsMiddleware.MetricsMiddleware)

	// Регистрируем роуты
	router.HandleFunc("/products/{id}", h.GetProduct).Methods(http.MethodGet)
	router.HandleFunc("/products", h.GetProducts).Methods(http.MethodGet)
	router.Handle("/products/{id}/add-to-cart",
		middleware.AuthMiddleware(http.HandlerFunc(h.AddProductToCart), h.grpcClient),
	).Methods(http.MethodPost)
	router.HandleFunc("/health", h.HealthCheck).Methods(http.MethodGet)

	// Prometheus metrics endpoint
	router.Handle("/metrics", promhttp.Handler())
}

// ---------- Helpers ----------
func writeJSON(w http.ResponseWriter, status int, payload any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(payload)
}

// ---------- Handlers ----------

func (h *Handler) GetProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	products, err := h.productStore.GetProducts(ctx)
	if err != nil {
		h.sugarLogger.Errorw("failed to get products", "error", err)
		http.Error(w, "Failed to get products", http.StatusInternalServerError)
		return
	}

	h.sugarLogger.Infow("products retrieved", "count", len(products))

	if err := writeJSON(w, http.StatusOK, products); err != nil {
		h.sugarLogger.Errorw("failed to write products response", "error", err)
	}
}

func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		if writeErr := writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid id"}); writeErr != nil {
			h.sugarLogger.Errorw("failed to write error response", "error", writeErr)
		}
		return
	}

	product, err := h.productStore.GetProductByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperrors.ErrNoProductFound) {
			if writeErr := writeJSON(w, http.StatusNotFound, map[string]any{"error": "product not found"}); writeErr != nil {
				h.sugarLogger.Errorw("failed to write error response", "error", writeErr)
			}
		} else {
			h.sugarLogger.Errorw("failed to get product", "error", err, "id", id)
			if writeErr := writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to load product"}); writeErr != nil {
				h.sugarLogger.Errorw("failed to write error response", "error", writeErr)
			}
		}
		return
	}

	// Записываем бизнес-метрику просмотра товара
	metrics.ProductViewsTotal.WithLabelValues(strconv.FormatInt(id, 10)).Inc()

	h.sugarLogger.Infow("product retrieved", "id", id)

	if err := writeJSON(w, http.StatusOK, product); err != nil {
		h.sugarLogger.Errorw("failed to write product response", "error", err)
	}
}

func (h *Handler) AddProductToCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userIDValue := r.Context().Value(middleware.UserIDKey)
	userID, ok := userIDValue.(int64)
	if !ok {
		h.sugarLogger.Errorw("invalid user_id type in context", "value", userIDValue)
		if writeErr := writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"}); writeErr != nil {
			h.sugarLogger.Errorw("failed to write error response", "error", writeErr)
		}
		return
	}

	vars := mux.Vars(r)
	productID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		if writeErr := writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid product id"}); writeErr != nil {
			h.sugarLogger.Errorw("failed to write error response", "error", writeErr)
		}
		return
	}

	product, err := h.productStore.GetProductByID(ctx, productID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNoProductFound) {
			if writeErr := writeJSON(w, http.StatusNotFound, map[string]any{"error": "product not found"}); writeErr != nil {
				h.sugarLogger.Errorw("failed to write error response", "error", writeErr)
			}
		} else {
			h.sugarLogger.Errorw("failed to load product for cart",
				"error", err, "product_id", productID, "user_id", userID)
			if writeErr := writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to load product"}); writeErr != nil {
				h.sugarLogger.Errorw("failed to write error response", "error", writeErr)
			}
		}
		return
	}

	_, err = h.cartStore.UpsertProductToCart(ctx, userID, product.ID, product.Price)
	if err != nil {
		h.sugarLogger.Errorw("failed to add product to cart",
			"error", err,
			"user_id", userID,
			"product_id", product.ID,
		)
		if writeErr := writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to add to cart"}); writeErr != nil {
			h.sugarLogger.Errorw("failed to write error response", "error", writeErr)
		}
		return
	}

	h.sugarLogger.Infow("product added to cart",
		"user_id", userID,
		"product_id", product.ID,
	)

	if err := writeJSON(w, http.StatusOK, map[string]any{
		"message": "product added to cart",
		"product": product,
	}); err != nil {
		h.sugarLogger.Errorw("failed to write success response", "error", err)
	}
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if err := writeJSON(w, http.StatusOK, map[string]any{
		"status":  "healthy",
		"service": "products-service",
	}); err != nil {
		h.sugarLogger.Errorw("failed to write health check response", "error", err)
	}
}
