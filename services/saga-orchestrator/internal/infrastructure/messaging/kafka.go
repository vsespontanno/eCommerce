package messaging

import (
	"context"
	"encoding/json"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/vsespontanno/eCommerce/services/saga-orchestrator/internal/domain/event/entity"
	"go.uber.org/zap"
)

type KafkaProducer struct {
	producer *kafka.Producer
	logger   *zap.SugaredLogger
	topic    string
}

func NewKafkaProducer(broker, topic, saslUsername, saslPassword, sslCAPath, securityProtocol, saslMechanism string, logger *zap.SugaredLogger) (*KafkaProducer, error) {
	config := &kafka.ConfigMap{
		"bootstrap.servers": broker,
		"acks":              "all",
		"retries":           10,
	}

	// Add SASL/SSL configuration if credentials are provided (for Yandex Cloud Kafka)
	if saslUsername != "" && saslPassword != "" {
		//nolint:errcheck // SetKey errors are non-critical for Kafka config
		_ = config.SetKey("security.protocol", securityProtocol)
		//nolint:errcheck
		_ = config.SetKey("sasl.mechanism", saslMechanism)
		//nolint:errcheck
		_ = config.SetKey("sasl.username", saslUsername)
		//nolint:errcheck
		_ = config.SetKey("sasl.password", saslPassword)

		if sslCAPath != "" {
			//nolint:errcheck
			_ = config.SetKey("ssl.ca.location", sslCAPath)
		}

		logger.Infow("Kafka producer configured with SASL/SSL",
			"security.protocol", securityProtocol,
			"sasl.mechanism", saslMechanism,
			"ssl.ca.location", sslCAPath,
		)
	} else {
		logger.Info("Kafka producer configured without SASL/SSL (local mode)")
	}

	kafkaProducer, err := kafka.NewProducer(config)
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
		if ev, ok := e.(*kafka.Message); ok {
			if ev.TopicPartition.Error != nil {
				k.logger.Errorw("Delivery failed",
					"error", ev.TopicPartition.Error,
					"key", string(ev.Key),
					"topic", *ev.TopicPartition.Topic,
				)
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
		k.logger.Errorw("Error marshaling message", "error", err, "stage: ", "ProccessEvent")
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
	k.logger.Infow("Message produced", "orderID", event.OrderID, "topic", k.topic)
	k.producer.Flush(5000) // ждём до 5 секунд доставки
	return nil
}

func (k *KafkaProducer) produce(msg *kafka.Message) error {
	err := k.producer.Produce(msg, nil)
	if err != nil {
		k.logger.Errorw("Error producing message", "error", err, "stage: ", "Produce")
		return err
	}
	k.logger.Infow("Message produced", "orderID", msg.Key, "topic", k.topic)
	return nil
}
