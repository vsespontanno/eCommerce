package outbox

import (
	"context"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

const (
	// flushTimeoutMs - таймаут ожидания доставки сообщения в Kafka (5 секунд)
	flushTimeoutMs = 5000
	// batchSize - максимальное количество событий за одну итерацию
	batchSize = 100
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
		`SELECT id, aggregate_id, event_type, payload
         FROM outbox
         WHERE status = 'pending'
         ORDER BY created_at
         LIMIT $1 FOR UPDATE SKIP LOCKED`, // SKIP LOCKED для конкурентности
		batchSize,
	)
	if err != nil {
		p.log.Errorw("failed to fetch outbox", "error", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var aggregateID string
		var eventType string
		var payload []byte

		if err = rows.Scan(&id, &aggregateID, &eventType, &payload); err != nil {
			p.log.Errorw("failed to scan outbox row", "error", err)
			continue
		}

		// Отправляем в Kafka с aggregate_id как ключом (для партиционирования по order_id)
		deliveryChan := make(chan kafka.Event, 1)
		err = p.producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &p.topic,
				Partition: kafka.PartitionAny,
			},
			Key:   []byte(aggregateID), // Используем order_id как ключ для партиционирования
			Value: payload,
		}, deliveryChan)

		if err != nil {
			p.log.Errorw("failed to produce message", "error", err, "id", id, "aggregateID", aggregateID)
			p.markAsFailed(ctx, id)
			continue
		}

		// КРИТИЧНО: Ждём подтверждения доставки ПЕРЕД обновлением статуса в БД
		// Это гарантирует at-least-once семантику
		remaining := p.producer.Flush(flushTimeoutMs)
		if remaining > 0 {
			p.log.Errorw("failed to flush message - timeout", "id", id, "aggregateID", aggregateID, "remaining", remaining)
			p.markAsFailed(ctx, id)
			continue
		}

		// Проверяем результат доставки
		select {
		case e := <-deliveryChan:
			m := e.(*kafka.Message)
			if m.TopicPartition.Error != nil {
				p.log.Errorw("kafka delivery failed",
					"error", m.TopicPartition.Error,
					"id", id,
					"aggregateID", aggregateID,
				)
				p.markAsFailed(ctx, id)
				continue
			}
			p.log.Debugw("kafka delivery confirmed",
				"id", id,
				"aggregateID", aggregateID,
				"partition", m.TopicPartition.Partition,
				"offset", m.TopicPartition.Offset,
			)
		default:
			// Сообщение было доставлено (Flush вернул 0)
		}

		// Помечаем как обработанное ТОЛЬКО после успешной доставки в Kafka
		_, err = p.db.ExecContext(ctx,
			"UPDATE outbox SET status = 'processed', processed_at = NOW() WHERE id = $1",
			id,
		)
		if err != nil {
			p.log.Errorw("failed to update outbox status to processed", "error", err, "id", id)
		} else {
			p.log.Infow("event published successfully",
				"id", id,
				"aggregateID", aggregateID,
				"eventType", eventType,
				"topic", p.topic,
			)
		}
	}

	if err = rows.Err(); err != nil {
		p.log.Errorw("error iterating outbox rows", "error", err)
	}
}

// markAsFailed помечает событие как failed в БД
func (p *Publisher) markAsFailed(ctx context.Context, id int64) {
	if _, err := p.db.ExecContext(ctx,
		"UPDATE outbox SET status = 'failed' WHERE id = $1", id,
	); err != nil {
		p.log.Errorw("failed to update outbox status to failed", "error", err, "id", id)
	}
}
