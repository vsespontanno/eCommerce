package repository

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/vsespontanno/eCommerce/order-service/internal/domain/order/entity"
	"go.uber.org/zap"
)

type OrderStore struct {
	db      *sqlx.DB
	builder sq.StatementBuilderType
	logger  *zap.SugaredLogger
}

func NewOrderStore(db *sqlx.DB, logger *zap.SugaredLogger) *OrderStore {
	return &OrderStore{
		db:      db,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
		logger:  logger,
	}
}

func (s *OrderStore) CreateOrder(ctx context.Context, order *entity.Order) (string, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// insert order
	var id string
	err = tx.QueryRowContext(ctx,
		`INSERT INTO orders (user_id, total, status) VALUES ($1, $2, $3) RETURNING id`,
		order.UserID, order.Total, order.Status,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("insert order: %w", err)
	}

	for _, it := range order.Items {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO order_items (order_id, product_id, price, quantity) VALUES ($1, $2, $3, $4)`,
			id, it.ProductID, it.Price, it.Quantity,
		)
		if err != nil {
			return "", fmt.Errorf("insert order_item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit: %w", err)
	}
	s.logger.Infow("order created", "order_id", id, "user_id", order.UserID)
	return id, nil
}

func (s *OrderStore) GetOrder(ctx context.Context, orderID string) (*entity.Order, error) {
	var o entity.Order
	err := s.db.GetContext(ctx, &o, `SELECT id, user_id, total, status, created_at FROM orders WHERE id = $1`, orderID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	rows, err := s.db.QueryxContext(ctx, `SELECT product_id, price, quantity FROM order_items WHERE order_id = $1`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var it entity.OrderItem
		if err := rows.StructScan(&it); err != nil {
			return nil, err
		}
		o.Items = append(o.Items, it)
	}
	return &o, nil
}

func (s *OrderStore) ListOrdersByUser(ctx context.Context, userID int64, limit, offset int) ([]entity.Order, error) {
	q := s.builder.Select("id, user_id, total, status, created_at").
		From("orders").
		Where(sq.Eq{"user_id": userID}).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).Offset(uint64(offset)).
		RunWith(s.db)

	rows, err := q.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	orders := make([]entity.Order, 0)
	for rows.Next() {
		var o entity.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Total, &o.Status, &o.CreatedAt); err != nil {
			return nil, err
		}
		// load items for each order (could be optimized)
		itemsRows, err := s.db.QueryxContext(ctx, `SELECT product_id, price, quantity FROM order_items WHERE order_id = $1`, o.ID)
		if err != nil {
			return nil, err
		}
		for itemsRows.Next() {
			var it entity.OrderItem
			if err := itemsRows.StructScan(&it); err == nil {
				o.Items = append(o.Items, it)
			}
		}
		itemsRows.Close()
		orders = append(orders, o)
	}
	return orders, nil
}
