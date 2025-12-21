package app

import (
	"github.com/vsespontanno/eCommerce/services/cart-service/internal/app/httpapp"
	"github.com/vsespontanno/eCommerce/services/cart-service/internal/application/cart"
	"go.uber.org/zap"
)

type App struct {
	HTTPApp *httpapp.App
	Service *cart.Service
}

func New(logger *zap.SugaredLogger, httpPort int, cartService *cart.Service) *App {
	httpApp := httpapp.New(httpPort, logger)
	return &App{
		HTTPApp: httpApp,
		Service: cartService,
	}
}
