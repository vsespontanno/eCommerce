package messaging

import (
	"context"
	"encoding/json"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/vsespontanno/eCommerce/cart-service/internal/domain/order/entity"
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

func NewKafkaConsumer(brokers, groupID, topic string, logger *zap.SugaredLogger, orderCompleter OrderCompleter) (*KafkaConsumer, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
	})
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
				if order.Status == "Pending" {
					if err := k.orderCompleter.CompleteOrder(ctx, order); err != nil {
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
