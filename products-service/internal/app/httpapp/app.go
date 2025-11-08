package httpapp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type App struct {
	log    *zap.SugaredLogger
	server *http.Server
	router *mux.Router
}

func New(port int, logger *zap.SugaredLogger) *App {
	router := mux.NewRouter()
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	return &App{
		log:    logger,
		server: httpServer,
		router: router,
	}
}

func (a *App) Run() error {
	a.log.Info("Starting HTTP server...")
	return a.server.ListenAndServe()
}

func (a *App) Shutdown(ctx context.Context) error {
	return a.server.Shutdown(ctx)
}

func (a *App) Router() *mux.Router {
	return a.router
}
