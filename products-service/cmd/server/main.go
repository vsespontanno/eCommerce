package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/vsespontanno/eCommerce/pkg/logger"
	"github.com/vsespontanno/eCommerce/products-service/internal/app"
	"github.com/vsespontanno/eCommerce/products-service/internal/client"
	"github.com/vsespontanno/eCommerce/products-service/internal/config"
	"github.com/vsespontanno/eCommerce/products-service/internal/domain/models"
	"github.com/vsespontanno/eCommerce/products-service/internal/handler"
	"github.com/vsespontanno/eCommerce/products-service/internal/repository"
	"github.com/vsespontanno/eCommerce/products-service/internal/repository/postgres"
	"github.com/vsespontanno/eCommerce/products-service/internal/service/saga"
)

func main() {
	logger.InitLogger()

	cfg, err := config.MustLoad()
	if err != nil {
		logger.Log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database (example: PostgreSQL)
	db, err := repository.ConnectToPostgres(
		cfg.PGUser,
		cfg.PGPassword,
		cfg.PGName,
		cfg.PGHost,
		cfg.PGPort,
	)
	if err != nil {
		logger.Log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	store := postgres.NewProductStore(db)
	cartStore := postgres.NewCartStore(db)
	sagaStore := postgres.NewSagaStore(db)
	sagaService := saga.NewSagaService(sagaStore, logger.Log)
	// Initialize application
	app := app.New(logger.Log, cfg.HTTPPort, cfg.GRPCProductsServerPort, cfg.GRPCSagaServerPort, store, sagaService)
	jwtClient := client.NewJwtClient(cfg.GRPCJwtPort)

	seedSomeValues(store)
	// Register handlers
	handler := handler.New(cartStore, store, logger.Log, jwtClient)
	handler.RegisterRoutes(app.HTTPApp.Router())

	// Start server in a goroutine
	go func() {
		if err := app.HTTPApp.Run(); err != nil {
			logger.Log.Errorf("HTTP server failed: %v", err)
		}
	}()

	go func() {
		if err := app.GRPCApp.Run(); err != nil {
			logger.Log.Errorf("gRPC server failed: %v", err)
		}
	}()

	// Handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	logger.Log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.HTTPApp.Shutdown(ctx); err != nil {
		logger.Log.Errorf("HTTP server shutdown failed: %v", err)
	} else {
		logger.Log.Info("HTTP Server gracefully stopped")
	}

	app.GRPCApp.Stop()
	logger.Log.Info("gRPC Server gracefully stopped")
	logger.Log.Info("Server stopped")
}

func seedSomeValues(store *postgres.ProductStore) {
	product1 := &models.Product{
		Name:         "Red Bull",
		Description:  "Good energy drink for gym",
		Price:        141.0,
		ID:           1,
		CountInStock: 100,
	}

	product2 := &models.Product{
		Name:         "Chapman Red",
		Description:  "vERy tasty ciagarettes for your deepression",
		Price:        253.0,
		ID:           2,
		CountInStock: 100,
	}

	store.SaveProduct(context.TODO(), product1)
	store.SaveProduct(context.TODO(), product2)
	fmt.Println("Products seeded")
}
