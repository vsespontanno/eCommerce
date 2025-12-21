package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/vsespontanno/eCommerce/pkg/logger"
	"github.com/vsespontanno/eCommerce/services/cart-service/internal/app"
	applicationCart "github.com/vsespontanno/eCommerce/services/cart-service/internal/application/cart"
	applicationOrder "github.com/vsespontanno/eCommerce/services/cart-service/internal/application/order"
	applicationSaga "github.com/vsespontanno/eCommerce/services/cart-service/internal/application/saga"
	"github.com/vsespontanno/eCommerce/services/cart-service/internal/config"
	jwtClient "github.com/vsespontanno/eCommerce/services/cart-service/internal/infrastructure/client/grpc/jwt"
	orderClient "github.com/vsespontanno/eCommerce/services/cart-service/internal/infrastructure/client/grpc/order"
	"github.com/vsespontanno/eCommerce/services/cart-service/internal/infrastructure/client/grpc/products"
	"github.com/vsespontanno/eCommerce/services/cart-service/internal/infrastructure/client/grpc/saga"
	"github.com/vsespontanno/eCommerce/services/cart-service/internal/infrastructure/db"
	"github.com/vsespontanno/eCommerce/services/cart-service/internal/infrastructure/jobs"
	"github.com/vsespontanno/eCommerce/services/cart-service/internal/infrastructure/messaging"
	"github.com/vsespontanno/eCommerce/services/cart-service/internal/infrastructure/repository/postgres"
	"github.com/vsespontanno/eCommerce/services/cart-service/internal/infrastructure/repository/redis"
	"github.com/vsespontanno/eCommerce/services/cart-service/internal/presentation/http/handlers"
	"github.com/vsespontanno/eCommerce/services/cart-service/internal/presentation/http/handlers/middleware"
)

func main() {
	logger.InitLogger()
	defer func() {
		if syncErr := logger.Log.Sync(); syncErr != nil {
			// Игнорируем ошибки sync для stderr/stdout
		}
	}()
	cfg, err := config.MustLoad()
	if err != nil {
		logger.Log.Fatalf("Failed to load config: %v", err)
	}

	ctx := context.TODO()
	// Initialize database (example: PostgreSQL)
	pg, err := db.ConnectToPostgres(
		cfg.PGUser,
		cfg.PGPassword,
		cfg.PGName,
		cfg.PGHost,
		cfg.PGPort,
	)
	if err != nil {
		logger.Log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pg.Close()
	// Initialize Redis with proper configuration
	redisAddr := cfg.RedisAddr
	if redisAddr == "" {
		redisAddr = "localhost:6379" // default
	}
	redisClient := db.ConnectToRedis(redisAddr, cfg.RedisPassword, cfg.RedisDB)
	defer redisClient.Close()

	productsClient := products.NewProductsClient(cfg.GRPCProductsClientPort, logger.Log)
	redisCleaner := redis.NewCleaner(redisClient, logger.Log)
	redisUpdater := redis.NewRedisUpdater(redisClient, logger.Log)
	sagaClient := saga.NewSagaClient(cfg.GRPCSagaClientPort, logger.Log)
	orderClient := orderClient.NewOrderClient(cfg.GRPCOrderClientPort, logger.Log)
	// Initialize cart service
	pgStore := postgres.NewCartStore(pg, logger.Log)
	redisStore := redis.NewCartStore(redisClient, logger.Log)
	cartService := applicationCart.NewCart(logger.Log, redisStore, productsClient, pgStore, cfg.MaxProductQuantity)
	sagaService := applicationSaga.NewSagaService(logger.Log, redisStore, sagaClient)
	rateLimiter := middleware.NewRateLimiter(redisClient, cfg.RateLimitRPS)
	orderService := applicationOrder.NewOrderCompleteService(logger.Log, pgStore, redisCleaner, orderClient)
	jobUpdater := jobs.NewCartSyncJob(pgStore, redisUpdater, logger.Log, time.Second*15)

	app := app.New(logger.Log, cfg.HTTPPort, cartService)
	grpcJWTClientPort := cfg.GRPCJWTClientPort
	jwtClient := jwtClient.NewJwtClient(grpcJWTClientPort)

	kafkaConsumer, err := messaging.NewKafkaConsumer(cfg.KafkaBroker, cfg.KafkaGroup, cfg.KafkaTopic, logger.Log, orderService)
	if err != nil {
		logger.Log.Fatalf("Failed to connect to kafka: %v", err)
	}

	// Создаем контекст для Kafka consumer с возможностью отмены
	kafkaCtx, kafkaCancel := context.WithCancel(context.Background())
	defer kafkaCancel()

	kafkaConsumer.Poll(kafkaCtx)

	handler := handlers.New(cartService, logger.Log, jwtClient, rateLimiter, sagaService)
	handler.RegisterRoutes(app.HTTPApp.Router())

	go func() {
		if err := app.HTTPApp.Run(); err != nil {
			logger.Log.Errorf("HTTP server failed: %v", err)
		}
	}()

	go jobUpdater.Start(ctx)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	logger.Log.Info("Shutting down server...")

	// Останавливаем Kafka consumer
	kafkaCancel()
	kafkaConsumer.Close()
	logger.Log.Info("Kafka consumer stopped")

	// Останавливаем HTTP сервер
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.HTTPApp.Shutdown(shutdownCtx); err != nil {
		logger.Log.Errorf("HTTP server shutdown failed: %v", err)
	} else {
		logger.Log.Info("HTTP Server gracefully stopped")
	}

}
