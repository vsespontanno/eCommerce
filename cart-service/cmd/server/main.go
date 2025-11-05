package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/vsespontanno/eCommerce/cart-service/internal/app"
	jwtClient "github.com/vsespontanno/eCommerce/cart-service/internal/client/jwt"
	"github.com/vsespontanno/eCommerce/cart-service/internal/client/products"
	"github.com/vsespontanno/eCommerce/cart-service/internal/config"
	"github.com/vsespontanno/eCommerce/cart-service/internal/handler"
	"github.com/vsespontanno/eCommerce/cart-service/internal/handler/middleware"
	"github.com/vsespontanno/eCommerce/cart-service/internal/repository"
	"github.com/vsespontanno/eCommerce/cart-service/internal/repository/postgres"
	"github.com/vsespontanno/eCommerce/cart-service/internal/repository/redis"
	"github.com/vsespontanno/eCommerce/cart-service/internal/service"
	"github.com/vsespontanno/eCommerce/pkg/logger"
)

func main() {
	logger.InitLogger()
	defer logger.Log.Sync()

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

	// Initialize Redis with proper configuration
	redisAddr := cfg.RedisAddr
	if redisAddr == "" {
		redisAddr = "localhost:6379" // default
	}
	redisClient := repository.ConnectToRedis(redisAddr, cfg.RedisPassword, cfg.RedisDB)
	defer redisClient.Close()

	productsClient := products.NewProductsClient(cfg.GRPCProductsClientPort, logger.Log)

	// Initialize cart service
	cartStore := postgres.NewCartStore(db)
	redisStore := redis.NewOrderStore(redisClient, logger.Log)
	cartService := service.NewCart(logger.Log, cartStore)
	orderService := service.NewOrder(logger.Log, redisStore, productsClient)

	// Initialize rate limiter
	rateLimiter := middleware.NewRateLimiter(redisClient, cfg.RateLimitRPS)

	// Initialize application
	app := app.New(logger.Log, cfg.HTTPPort, cartService)
	grpcJWTClientPort := cfg.GRPCJWTClientPort
	if grpcJWTClientPort == "" {
		grpcJWTClientPort = "50051" // default
	}
	jwtClient := jwtClient.NewJwtClient(grpcJWTClientPort)

	// Register handlers
	handler := handler.New(cartService, logger.Log, jwtClient, orderService, rateLimiter)
	handler.RegisterRoutes(app.HTTPApp.Router())

	// Start server in a goroutine
	go func() {
		if err := app.HTTPApp.Run(); err != nil {
			logger.Log.Errorf("HTTP server failed: %v", err)
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

}
