package handler

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Handler struct {
}

func New() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	//it would be in potential , but know for beginning we create simple thing
	// router.HandleFunc("/products", h.GetProducts).Methods(http.MethodGet)
	// router.HandleFunc("/products/{id}", h.GetProduct).Methods(http.MethodGet)
	// router.HandleFunc("/products", h.CreateProduct).Methods(http.MethodPost)

	router.HandleFunc("/products", h.GetProducts).Methods(http.MethodGet)
}

func (h *Handler) GetProducts(w http.ResponseWriter, r *http.Request) {
	//here would be the logic to retrieve products but know just testing
	w.Write([]byte("List of products"))
}
