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
	"github.com/vsespontanno/eCommerce/cart-service/internal/handler/middleware"
	"github.com/vsespontanno/eCommerce/cart-service/internal/repository"
	"github.com/vsespontanno/eCommerce/cart-service/internal/repository/postgres"
	"github.com/vsespontanno/eCommerce/cart-service/internal/repository/redis"
	"github.com/vsespontanno/eCommerce/cart-service/internal/service"
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

	// Initialize Redis with proper configuration
	redisAddr := cfg.RedisAddr
	if redisAddr == "" {
		redisAddr = "localhost:6379" // default
	}
	redisClient := repository.ConnectToRedis(redisAddr, cfg.RedisPassword, cfg.RedisDB)
	defer redisClient.Close()

	// Initialize cart service
	cartStore := postgres.NewCartStore(db)
	redisStore := redis.NewOrderStore(redisClient)
	cartService := service.NewCart(logger.Sugar(), cartStore)
	orderService := service.NewOrder(logger.Sugar(), redisStore)

	// Initialize rate limiter
	rateLimiter := middleware.NewRateLimiter(redisClient, cfg.RateLimitRPS)

	// Initialize application
	app := app.New(*logger, cfg.HTTPPort, cartService)
	grpcClientPort := cfg.GRPCPort
	if grpcClientPort == "" {
		grpcClientPort = "50051" // default
	}
	jwtClient := client.NewJwtClient(grpcClientPort)

	// Register handlers
	handler := handler.New(cartService, sugar, jwtClient, orderService, rateLimiter)
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
