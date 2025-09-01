package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/vsespontanno/eCommerce/cart-service/internal/app"
	"github.com/vsespontanno/eCommerce/cart-service/internal/client"
	"github.com/vsespontanno/eCommerce/cart-service/internal/config"
	"github.com/vsespontanno/eCommerce/cart-service/internal/handler"
	"github.com/vsespontanno/eCommerce/cart-service/internal/repository"
	"github.com/vsespontanno/eCommerce/cart-service/internal/repository/postgres"
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

	cartStore := postgres.NewCartStore(db)
	// Initialize application
	app := app.New(*logger, cfg.HTTPPort, cartStore)
	grpcClientPort := "50051"
	jwtClient := client.NewJwtClient(grpcClientPort)

	// Register handlers
	handler := handler.New(cartStore, sugar, jwtClient)
	handler.RegisterRoutes(app.HTTPApp.Router())

	// Start server in a goroutine
	go func() {
		if err := app.HTTPApp.Run(); err != nil {
			sugar.Errorf("HTTP server failed: %v", err)
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

}
