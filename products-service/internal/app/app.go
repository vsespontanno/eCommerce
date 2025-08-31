package app

import (
	"github.com/vsespontanno/eCommerce/products-service/internal/app/grpcapp"
	"github.com/vsespontanno/eCommerce/products-service/internal/app/httpapp"
	"github.com/vsespontanno/eCommerce/products-service/internal/repository/postgres"
	"go.uber.org/zap"
)

type App struct {
	HTTPApp *httpapp.App
	GRPCApp *grpcapp.App
	Store   *postgres.ProductStore
}

func New(logger zap.Logger, httpPort int, grpcPort int, store *postgres.ProductStore) *App {
	httpApp := httpapp.New(httpPort, &logger)
	grpcApp := grpcapp.NewApp(logger.Sugar(), store, grpcPort)
	return &App{
		HTTPApp: httpApp,
		GRPCApp: grpcApp,
		Store:   store,
	}
}
