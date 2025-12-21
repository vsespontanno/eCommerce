package outbox

import (
	"context"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type Publisher struct {
	db       *sqlx.DB
	producer *kafka.Producer
	interval time.Duration
	log      *zap.SugaredLogger
	topic    string
}

func NewOutboxPublisher(db *sqlx.DB, producer *kafka.Producer, log *zap.SugaredLogger, topic string, interval time.Duration) *Publisher {
	return &Publisher{
		db:       db,
		producer: producer,
		interval: interval,
		log:      log,
		topic:    topic,
	}
}

func (p *Publisher) Start(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	p.log.Infow("outbox publisher started", "interval", p.interval, "topic", p.topic)

	for {
		select {
		case <-ctx.Done():
			p.log.Info("outbox publisher stopped")
			return
		case <-ticker.C:
			p.processOutbox(ctx)
		}
	}
}

func (p *Publisher) processOutbox(ctx context.Context) {
	rows, err := p.db.QueryContext(ctx,
		`SELECT id, event_type, payload 
         FROM outbox 
         WHERE status = 'pending' 
         ORDER BY created_at 
         LIMIT 100 FOR UPDATE SKIP LOCKED`, // SKIP LOCKED для конкурентности
	)
	if err != nil {
		p.log.Errorw("failed to fetch outbox", "error", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var eventType string
		var payload []byte

		if err = rows.Scan(&id, &eventType, &payload); err != nil {
			continue
		}

		// 2. Отправляем в Kafka
		err = p.producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &p.topic,
				Partition: kafka.PartitionAny,
			},
			Key:   []byte(eventType),
			Value: payload,
		}, nil)

		if err != nil {
			p.log.Errorw("failed to produce message", "error", err, "id", id)
			if _, execErr := p.db.ExecContext(ctx,
				"UPDATE outbox SET status = 'failed' WHERE id = $1", id,
			); execErr != nil {
				p.log.Errorw("failed to update outbox status to failed", "error", execErr, "id", id)
			}
			continue
		}

		// Помечаем как обработанное
		_, err = p.db.ExecContext(ctx,
			"UPDATE outbox SET status = 'processed', processed_at = NOW() WHERE id = $1",
			id,
		)
		if err != nil {
			p.log.Errorw("failed to update outbox", "error", err, "id", id)
		} else {
			p.log.Infow("event published successfully", "id", id, "eventType", eventType, "topic", p.topic)
		}
	}
}
