package app

import (
	"github.com/vsespontanno/eCommerce/cart-service/internal/app/httpapp"
	"github.com/vsespontanno/eCommerce/cart-service/internal/repository/postgres"
	"go.uber.org/zap"
)

type App struct {
	HTTPApp *httpapp.App
	Store   *postgres.CartStore
}

func New(logger zap.Logger, httpPort int, store *postgres.CartStore) *App {
	httpApp := httpapp.New(httpPort, &logger)
	return &App{
		HTTPApp: httpApp,
		Store:   store,
	}
}
