package messaging

import (
	"context"
	"encoding/json"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/vsespontanno/eCommerce/services/cart-service/internal/domain/order/entity"
	"go.uber.org/zap"
)

type OrderCompleter interface {
	CompleteOrder(ctx context.Context, order *entity.OrderEvent) error
}

type KafkaConsumer struct {
	consumer       *kafka.Consumer
	topic          string
	logger         *zap.SugaredLogger
	orderCompleter OrderCompleter
}

func NewKafkaConsumer(brokers, groupID, topic, saslUsername, saslPassword, sslCAPath, securityProtocol, saslMechanism string, logger *zap.SugaredLogger, orderCompleter OrderCompleter) (*KafkaConsumer, error) {
	config := &kafka.ConfigMap{
		"bootstrap.servers": brokers,
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
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

		logger.Infow("Kafka consumer configured with SASL/SSL",
			"security.protocol", securityProtocol,
			"sasl.mechanism", saslMechanism,
			"ssl.ca.location", sslCAPath,
		)
	} else {
		logger.Info("Kafka consumer configured without SASL/SSL (local mode)")
	}

	c, err := kafka.NewConsumer(config)
	if err != nil {
		return nil, err
	}

	if err := c.SubscribeTopics([]string{topic}, nil); err != nil {
		return nil, err
	}

	return &KafkaConsumer{
		consumer:       c,
		topic:          topic,
		logger:         logger,
		orderCompleter: orderCompleter,
	}, nil
}

func (k *KafkaConsumer) Close() {
	k.consumer.Close()
}

func (k *KafkaConsumer) Poll(ctx context.Context) {
	go func() {
		k.logger.Info("Kafka consumer started", "topic ", k.topic)
		for {
			select {
			case <-ctx.Done():
				k.logger.Info("Kafka consumer stopped", " topic ", k.topic)
				return
			default:
				msg, err := k.consumer.ReadMessage( /*100 * time.Millisecond*/ -1)
				if err != nil {
					if kafkaErr, ok := err.(kafka.Error); ok && kafkaErr.Code() == kafka.ErrTimedOut {
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
					// Ошибка парсинга - коммитим offset чтобы не застрять на битом сообщении
					k.logger.Errorw("Failed to parse message, committing offset to skip",
						"error", err,
						"partition", msg.TopicPartition.Partition,
						"offset", msg.TopicPartition.Offset,
					)
					if _, commitErr := k.consumer.CommitMessage(msg); commitErr != nil {
						k.logger.Errorw("Error committing offset after parse error", "error", commitErr)
					}
					continue
				}

				// Обрабатываем только успешно завершенные заказы из saga
				if order.Status == "Completed" {
					if err := k.orderCompleter.CompleteOrder(ctx, order); err != nil {
						k.logger.Errorw("Error completing order",
							"order_id", order.OrderID,
							"user_id", order.UserID,
							"eventType", order.EventType,
							"error", err,
						)
						// НЕ коммитим offset при ошибке - сообщение будет обработано повторно
						continue
					}
					k.logger.Infow("Order completed successfully",
						"order_id", order.OrderID,
						"user_id", order.UserID,
						"eventType", order.EventType,
						"total", order.Total,
					)
				} else {
					k.logger.Warnw("Received order with unexpected status",
						"order_id", order.OrderID,
						"status", order.Status,
						"eventType", order.EventType,
					)
				}

				// Коммитим offset только после успешной обработки
				if _, err := k.consumer.CommitMessage(msg); err != nil {
					k.logger.Errorw("Error committing offset", "error", err, "order_id", order.OrderID)
				}

			}
		}
	}()
}

func (k *KafkaConsumer) processMessage(msg *kafka.Message) (*entity.OrderEvent, error) {
	var cartOrder entity.OrderEvent
	err := json.Unmarshal(msg.Value, &cartOrder)
	k.logger.Infof("got msg: %+v", cartOrder)
	if err != nil {
		k.logger.Errorw("Error unmarshalling message", "error", err, "stage: ", "processMessage")
		return &entity.OrderEvent{}, err
	}
	return &cartOrder, nil
}
