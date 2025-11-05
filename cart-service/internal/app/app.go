package app

import (
	"github.com/vsespontanno/eCommerce/cart-service/internal/app/httpapp"
	"github.com/vsespontanno/eCommerce/cart-service/internal/service"
	"go.uber.org/zap"
)

type App struct {
	HTTPApp *httpapp.App
	Service *service.CartService
}

func New(logger *zap.SugaredLogger, httpPort int, cartService *service.CartService) *App {
	httpApp := httpapp.New(httpPort, logger)
	return &App{
		HTTPApp: httpApp,
		Service: cartService,
	}
}
