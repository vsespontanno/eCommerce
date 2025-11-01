package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/vsespontanno/eCommerce/cart-service/internal/domain/models"
	"go.uber.org/zap"
)

type CartChanger interface {
}

type KafkaConsumer struct {
	topic    string
	groupID  string
	consumer *kafka.Consumer
	logger   *zap.SugaredLogger
}

func NewKafkaConsumer(broker string, topic string, groupID string, logger *zap.SugaredLogger) (*KafkaConsumer, error) {
	kafkaConsumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": broker,
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
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
		select {
		case <-ctx.Done():
			k.logger.Info("Kafka consumer stopped", "topic", k.topic)
			return
		default:
			msg, err := k.consumer.ReadMessage(-1)
			if err == nil {
				k.logger.Infow("Received message", "topic", *msg.TopicPartition.Topic, "partition", msg.TopicPartition.Partition, "offset", msg.TopicPartition.Offset,
					"key", string(msg.Key), "value", string(msg.Value),
				)
				order, err := k.processMessage(msg)
				if err == nil {
					doSomethingWithOrder(order)
				}
			}
		}
	}()
}

func (k *KafkaConsumer) processMessage(msg *kafka.Message) (models.SagaCart, error) {
	var cartOrder models.SagaCart
	err := json.Unmarshal(msg.Value, &cartOrder)
	if err != nil {
		k.logger.Errorw("Error unmarshalling message", "error", err, "stage: ", "processMessage")
		return models.SagaCart{}, err
	}
	return cartOrder, nil
}

func doSomethingWithOrder(cartOrder models.SagaCart) {
	fmt.Println("Doing some stuff")
}
