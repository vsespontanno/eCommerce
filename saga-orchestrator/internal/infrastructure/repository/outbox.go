package repository

import (
	"context"
	"encoding/json"

	"github.com/jmoiron/sqlx"
	"github.com/vsespontanno/eCommerce/saga-orchestrator/internal/domain/event/entity"
	"go.uber.org/zap"
)

type OutboxRepository struct {
	db  *sqlx.DB
	log *zap.SugaredLogger
}

func NewOutboxRepository(db *sqlx.DB, log *zap.SugaredLogger) *OutboxRepository {
	return &OutboxRepository{
		db:  db,
		log: log,
	}
}

// SaveEvent сохраняет событие в outbox таблицу
func (r *OutboxRepository) SaveEvent(ctx context.Context, event entity.OrderEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		r.log.Errorw("failed to marshal event", "error", err, "orderID", event.OrderID)
		return err
	}

	query := `
		INSERT INTO outbox (aggregate_id, aggregate_type, event_type, payload, status)
		VALUES ($1, $2, $3, $4, 'pending')
	`

	_, err = r.db.ExecContext(ctx, query,
		event.OrderID,
		"saga",
		"OrderCompleted",
		payload,
	)

	if err != nil {
		r.log.Errorw("failed to save event to outbox", "error", err, "orderID", event.OrderID)
		return err
	}

	r.log.Infow("event saved to outbox", "orderID", event.OrderID, "eventType", "OrderCompleted")
	return nil
}
