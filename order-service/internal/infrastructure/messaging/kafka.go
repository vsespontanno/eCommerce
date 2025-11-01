package messaging

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/zap"
)

type KafkaProducer struct {
	producer *kafka.Producer
	logger   *zap.SugaredLogger
}

func NewKafkaProducer(broker string, topic string, groupID string, logger *zap.SugaredLogger) (*KafkaProducer, error) {
	kafkaProducer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": broker,
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		logger.Errorw("Error creating kafka producer", "error", err, "stage: ", "NewKafkaProducer")
		return nil, err
	}
	return &KafkaProducer{
		producer: kafkaProducer,
		logger:   logger,
	}, nil
}

func (k *KafkaProducer) Close() {
	k.producer.Close()
}

func (k *KafkaProducer) Produce(msg *kafka.Message) error {
	err := k.producer.Produce(msg, nil)
	if err != nil {
		k.logger.Errorw("Error producing message", "error", err, "stage: ", "Produce")
		return err
	}
	return nil
}
