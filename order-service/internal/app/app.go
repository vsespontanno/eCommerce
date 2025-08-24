package app

import (
	"github.com/vsespontanno/eCommerce/purchase-service/internal/app/httpapp"
	"go.uber.org/zap"
)

type App struct {
	HTTPApp *httpapp.App
	Store   *postgres.ProductStore
}

func New(logger zap.Logger, httpPort int, store *postgres.ProductStore) *App {
	httpApp := httpapp.New(httpPort, &logger)
	return &App{
		HTTPApp: httpApp,
		Store:   store,
	}
}
