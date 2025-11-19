package app

import (
	"github.com/vsespontanno/eCommerce/cart-service/internal-new/app/httpapp"
	"github.com/vsespontanno/eCommerce/cart-service/internal-new/application/cart"
	"go.uber.org/zap"
)

type App struct {
	HTTPApp *httpapp.App
	Service *cart.CartService
}

func New(logger *zap.SugaredLogger, httpPort int, cartService *cart.CartService) *App {
	httpApp := httpapp.New(httpPort, logger)
	return &App{
		HTTPApp: httpApp,
		Service: cartService,
	}
}
