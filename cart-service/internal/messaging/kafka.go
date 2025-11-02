package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/vsespontanno/eCommerce/cart-service/internal/domain/models"
	"go.uber.org/zap"
)

type CartCleaner interface {
	CleanCart(ctx context.Context, order models.OrderEvent) error
}

type KafkaConsumer struct {
	topic    string
	groupID  string
	consumer *kafka.Consumer
	logger   *zap.SugaredLogger
	Cleaner  CartCleaner
}

func NewKafkaConsumer(broker string, topic string, groupID string, logger *zap.SugaredLogger, cleaner CartCleaner) (*KafkaConsumer, error) {
	kafkaConsumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  broker,
		"group.id":           groupID,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": false,
	})
	if err != nil {
		logger.Errorw("Error creating kafka consumer", "error", err, "stage: ", "NewKafkaConsumer")
		return nil, err
	}
	return &KafkaConsumer{
		topic:    topic,
		groupID:  groupID,
		consumer: kafkaConsumer,
		logger:   logger,
		Cleaner:  cleaner,
	}, nil
}

func (k *KafkaConsumer) Close() {
	k.consumer.Close()
}

func (k *KafkaConsumer) Subscribe() error {
	err := k.consumer.Subscribe(k.topic, nil)
	if err != nil {
		k.logger.Errorw("Error subscribing to kafka topic", "error", err, "stage: ", "Subscribe")
		return err
	}
	return nil
}

func (k *KafkaConsumer) Poll(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				k.logger.Info("Kafka consumer stopped", "topic", k.topic)
				return
			default:
				msg, err := k.consumer.ReadMessage(100 * time.Millisecond)
				if err != nil {
					if err.(kafka.Error).Code() == kafka.ErrTimedOut {
						continue
					}
					k.logger.Errorw("Error reading message", "error", err)
					continue
				}
				k.logger.Infow("Received message",
					"topic", *msg.TopicPartition.Topic,
					"partition", msg.TopicPartition.Partition,
					"offset", msg.TopicPartition.Offset,
					"key", string(msg.Key),
					"value", string(msg.Value),
				)

				order, err := k.processMessage(msg)

				if err != nil {
					continue
				}
				if order.Status == "completed" {
					if err := k.Cleaner.CleanCart(ctx, order); err != nil {
						k.logger.Errorw("Error cleaning cart", "order_id", order.OrderID, "error", err)
						continue
					}
				}

				if _, err := k.consumer.CommitMessage(msg); err != nil {
					k.logger.Errorw("Error committing offset", "error", err)
				}

			}
		}
	}()
}

func (k *KafkaConsumer) processMessage(msg *kafka.Message) (models.OrderEvent, error) {
	var cartOrder models.OrderEvent
	err := json.Unmarshal(msg.Value, &cartOrder)
	if err != nil {
		k.logger.Errorw("Error unmarshalling message", "error", err, "stage: ", "processMessage")
		return models.OrderEvent{}, err
	}
	return cartOrder, nil
}

func doSomethingWithOrder(cartOrder models.OrderEvent) {
	fmt.Println("Doing some stuff")
}
