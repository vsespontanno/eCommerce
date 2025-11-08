package main

import (
	"context"

	"github.com/vsespontanno/eCommerce/order-service/internal/config"
	"github.com/vsespontanno/eCommerce/order-service/internal/domain/event/entity"
	productEntity "github.com/vsespontanno/eCommerce/order-service/internal/domain/product/entity"
	"github.com/vsespontanno/eCommerce/order-service/internal/infrastructure/messaging"
	"github.com/vsespontanno/eCommerce/pkg/logger"
)

func main() {
	cfg, err := config.MustLoad()
	if err != nil {
		panic(err)
	}
	logger.InitLogger()
	kafkaProducer, err := messaging.NewKafkaProducer(cfg.KafkaBroker, cfg.KafkaTopic, logger.Log)
	if err != nil {
		logger.Log.Errorw("error while creating kafka producer ", "error", err, "stage: ", "order. main")
	}
	exampleMessage1 := entity.OrderEvent{
		OrderID: "order-1001",
		UserID:  42,
		Products: []productEntity.Product{
			{ID: 1, Quantity: 2},
			{ID: 3, Quantity: 1},
		},
		Total:  4500,
		Status: "completed",
	}

	err = kafkaProducer.ProccessEvent(context.Background(), exampleMessage1)
	if err != nil {
		logger.Log.Errorw("error while publishing kafka message ", "error", err, "stage: ", "order. main")
		return
	}
}
