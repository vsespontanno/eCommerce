package messaging

import (
	"context"
	"encoding/json"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/vsespontanno/eCommerce/order-service/internal/domain/event/entity"
	"go.uber.org/zap"
)

type KafkaProducer struct {
	producer *kafka.Producer
	logger   *zap.SugaredLogger
	topic    string
}

func NewKafkaProducer(broker string, topic string, logger *zap.SugaredLogger) (*KafkaProducer, error) {
	kafkaProducer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": broker,
		"acks":              "all",
		"retries":           10,
	})
	if err != nil {
		logger.Errorw("Error creating kafka producer", "error", err, "stage: ", "NewKafkaProducer")
		return nil, err
	}
	kp := &KafkaProducer{
		producer: kafkaProducer,
		logger:   logger,
		topic:    topic,
	}

	go kp.monitorDelivery()

	return kp, nil
}

func (k *KafkaProducer) monitorDelivery() {
	for e := range k.producer.Events() {
		switch ev := e.(type) {
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				k.logger.Errorw("Delivery failed",
					"error", ev.TopicPartition.Error,
					"key", string(ev.Key),
					"topic", *ev.TopicPartition.Topic,
				)
				// TODO: retry или dead-letter topic
			} else {
				k.logger.Debugw("Message delivered",
					"key", string(ev.Key),
					"topic", *ev.TopicPartition.Topic,
					"partition", ev.TopicPartition.Partition,
					"offset", ev.TopicPartition.Offset,
				)
			}
		}
	}
}

func (k *KafkaProducer) Close() {
	k.producer.Flush(15 * 1000)
	k.producer.Close()
}

func (k *KafkaProducer) ProccessEvent(ctx context.Context, event entity.OrderEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		k.logger.Errorw("Error marshalling message", "error", err, "stage: ", "ProccessEvent")
		return err
	}
	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &k.topic, Partition: kafka.PartitionAny},
		Key:            []byte(event.OrderID),
		Value:          data,
		Timestamp:      time.Now(),
	}
	err = k.produce(msg)
	if err != nil {
		k.logger.Errorw("Error producing message", "error", err, "stage: ", "ProccessEvent")
		return err
	}
	return nil
}

func (k *KafkaProducer) produce(msg *kafka.Message) error {
	err := k.producer.Produce(msg, nil)
	if err != nil {
		k.logger.Errorw("Error producing message", "error", err, "stage: ", "Produce")
		return err
	}
	return nil
}
