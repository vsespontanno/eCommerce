package app

import (
	"github.com/vsespontanno/eCommerce/products-service/internal/app/grpcapp"
	"github.com/vsespontanno/eCommerce/products-service/internal/app/httpapp"
	"github.com/vsespontanno/eCommerce/products-service/internal/grpc/saga"
	"github.com/vsespontanno/eCommerce/products-service/internal/repository/postgres"
	"go.uber.org/zap"
)

type App struct {
	HTTPApp *httpapp.App
	GRPCApp *grpcapp.App
	Store   *postgres.ProductStore
}

func New(logger *zap.SugaredLogger, httpPort int, grpcProductsPort int, grpcSagaPort int, store *postgres.ProductStore, reserver saga.Reserver) *App {
	httpApp := httpapp.New(httpPort, logger)
	grpcApp := grpcapp.NewApp(logger, store, reserver, grpcProductsPort, grpcSagaPort)
	return &App{
		HTTPApp: httpApp,
		GRPCApp: grpcApp,
		Store:   store,
	}
}
