package messaging

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type CartChanger interface {
}

type KafkaConsumer struct {
}

// TODO: implement; function stub
func (kc *KafkaConsumer) Subscribe(topic string) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{})
	if err != nil {
		return
	}
	_ = c
}
