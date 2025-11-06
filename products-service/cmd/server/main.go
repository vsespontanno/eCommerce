package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/vsespontanno/eCommerce/products-service/internal/app"
	"github.com/vsespontanno/eCommerce/products-service/internal/client"
	"github.com/vsespontanno/eCommerce/products-service/internal/config"
	"github.com/vsespontanno/eCommerce/products-service/internal/domain/models"
	"github.com/vsespontanno/eCommerce/products-service/internal/handler"
	"github.com/vsespontanno/eCommerce/products-service/internal/repository"
	"github.com/vsespontanno/eCommerce/products-service/internal/repository/postgres"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	cfg, err := config.MustLoad()
	if err != nil {
		sugar.Fatalf("Failed to load config: %v", err)
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
		sugar.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	store := postgres.NewProductStore(db)
	cartStore := postgres.NewCartStore(db)
	// Initialize application
	app := app.New(*logger, cfg.HTTPPort, cfg.GRPCServerPort, store)
	jwtClient := client.NewJwtClient(cfg.GRPCJwtPort)

	// Register handlers
	handler := handler.New(cartStore, store, sugar, jwtClient)
	handler.RegisterRoutes(app.HTTPApp.Router())
	seedSomeValues(store)

	// Start server in a goroutine
	go func() {
		if err := app.HTTPApp.Run(); err != nil {
			sugar.Errorf("HTTP server failed: %v", err)
		}
	}()

	go func() {
		if err := app.GRPCApp.Run(); err != nil {
			sugar.Errorf("gRPC server failed: %v", err)
		}
	}()

	// Handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	sugar.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.HTTPApp.Shutdown(ctx); err != nil {
		sugar.Errorf("HTTP server shutdown failed: %v", err)
	} else {
		sugar.Info("HTTP Server gracefully stopped")
	}

	app.GRPCApp.Stop()
	sugar.Info("gRPC Server gracefully stopped")
	sugar.Info("Server stopped")
}

func seedSomeValues(store *postgres.ProductStore) {
	product1 := &models.Product{
		Name:        "Red Bull",
		Description: "Good energy drink for gym",
		Price:       141.0,
		ID:          1,
	}

	product2 := &models.Product{
		Name:        "Chapman Red",
		Description: "vERy tasty ciagarettes for your deepression",
		Price:       253.0,
		ID:          2,
	}

	store.SaveProduct(context.TODO(), product1)
	store.SaveProduct(context.TODO(), product2)
	fmt.Println("Products seeded")
}
