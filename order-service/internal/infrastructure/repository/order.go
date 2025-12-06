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

// ============= CREATE ORDER =================

func (s *OrderStore) CreateOrder(ctx context.Context, order *entity.Order) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("tx begin: %w", err)
	}
	defer tx.Rollback()

	// Проверка идемпотентности - заказ с таким ID уже существует?
	var exists bool
	err = tx.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM orders WHERE id = $1)`,
		order.OrderID,
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check order exists: %w", err)
	}

	if exists {
		s.logger.Infow("order already exists, skipping", "order_id", order.OrderID)
		return nil // Идемпотентность - заказ уже создан
	}

	// insert into orders (UUID as string)
	_, err = tx.ExecContext(ctx,
		`INSERT INTO orders (id, user_id, total, status)
         VALUES ($1, $2, $3, $4)`,
		order.OrderID, order.UserID, order.Total, order.Status,
	)
	if err != nil {
		return fmt.Errorf("insert order: %w", err)
	}

	for _, it := range order.Products {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO order_items (order_id, product_id, quantity)
             VALUES ($1, $2, $3)`,
			order.OrderID, it.ProductID, it.Quantity,
		)
		if err != nil {
			return fmt.Errorf("insert item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	s.logger.Infow("order saved", "order_id", order.OrderID)
	return nil
}

// ============= GET ORDER =====================

func (s *OrderStore) GetOrder(ctx context.Context, id string) (*entity.Order, error) {
	var o entity.Order

	err := s.db.GetContext(ctx, &o,
		`SELECT id AS order_id, user_id, total, status, created_at
         FROM orders WHERE id = $1`, id,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("select order: %w", err)
	}

	rows, err := s.db.QueryxContext(ctx,
		`SELECT product_id, quantity
         FROM order_items WHERE order_id = $1`, id,
	)
	if err != nil {
		return nil, fmt.Errorf("select order items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var it entity.OrderItem
		if err := rows.StructScan(&it); err != nil {
			return nil, err
		}
		o.Products = append(o.Products, it)
	}

	return &o, nil
}

// ============= LIST ORDERS ====================

func (s *OrderStore) ListOrdersByUser(ctx context.Context, userID int64, limit, offset int) ([]entity.Order, error) {
	q := s.builder.
		Select("id AS order_id", "user_id", "total", "status").
		From("orders").
		Where(sq.Eq{"user_id": userID}).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		RunWith(s.db)

	rows, err := q.QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("query orders: %w", err)
	}
	defer rows.Close()

	orders := make([]entity.Order, 0)

	for rows.Next() {
		var o entity.Order
		if err := rows.Scan(&o.OrderID, &o.UserID, &o.Total, &o.Status); err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}

		// load items
		itemsRows, err := s.db.QueryxContext(ctx,
			`SELECT product_id, quantity 
             FROM order_items WHERE order_id = $1`, o.OrderID,
		)
		if err != nil {
			return nil, fmt.Errorf("items: %w", err)
		}

		for itemsRows.Next() {
			var it entity.OrderItem
			if err := itemsRows.StructScan(&it); err == nil {
				o.Products = append(o.Products, it)
			}
		}
		itemsRows.Close()

		orders = append(orders, o)
	}

	return orders, nil
}
