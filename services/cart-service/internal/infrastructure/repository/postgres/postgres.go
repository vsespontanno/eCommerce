package postgres

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/vsespontanno/eCommerce/services/cart-service/internal/domain/apperrors"
	"github.com/vsespontanno/eCommerce/services/cart-service/internal/domain/cart/entity"
	orderEntity "github.com/vsespontanno/eCommerce/services/cart-service/internal/domain/order/entity"
	"go.uber.org/zap"
)

type CartStore struct {
	db      *sqlx.DB
	logger  *zap.SugaredLogger
	builder sq.StatementBuilderType
}

func NewCartStore(db *sqlx.DB, logger *zap.SugaredLogger) *CartStore {
	return &CartStore{
		db:      db,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
		logger:  logger,
	}
}

func (s *CartStore) GetCart(ctx context.Context, userID int64) (*entity.Cart, error) {
	var cart entity.Cart
	query := s.builder.
		Select("user_id, product_id, quantity, amount_for_product").
		From("cart").
		Where(sq.Eq{"user_id": userID}).
		RunWith(s.db)

	rows, err := query.QueryContext(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return &entity.Cart{}, apperrors.ErrNoCartFound
		}
		return &entity.Cart{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var item entity.CartItem
		if err := rows.Scan(&item.UserID, &item.ProductID, &item.Quantity, &item.Price); err != nil {
			return &entity.Cart{}, err
		}
		cart.Items = append(cart.Items, item)
	}

	// Проверяем ошибки, возникшие во время итерации
	if err := rows.Err(); err != nil {
		return &entity.Cart{}, err
	}

	if len(cart.Items) == 0 {
		return &entity.Cart{}, apperrors.ErrNoCartFound
	}

	return &cart, nil
}

func (s *CartStore) UpsertCart(ctx context.Context, userID int64, cart *[]entity.CartItem) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		_ = tx.Rollback()
	}()

	for _, item := range *cart {
		qb := s.builder.
			Insert("cart").
			Columns("user_id", "product_id", "quantity", "amount_for_product").
			Values(userID, item.ProductID, item.Quantity, item.Price).
			Suffix(`
				ON CONFLICT (user_id, product_id)
				DO UPDATE SET quantity = EXCLUDED.quantity
			`)

		sqlStr, args, err := qb.ToSql()
		if err != nil {
			return fmt.Errorf("failed to build SQL: %w", err)
		}

		if _, err := tx.ExecContext(ctx, sqlStr, args...); err != nil {
			s.logger.Errorw("failed to upsert cart item",
				"user_id", userID,
				"product_id", item.ProductID,
				"error", err,
			)
			return fmt.Errorf("failed to exec upsert: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit upsert: %w", err)
	}

	s.logger.Infow("cart upserted successfully", "user_id", userID, "items", len(*cart))
	return nil
}

func (s *CartStore) CleanCart(ctx context.Context, order *orderEntity.OrderEvent) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		s.logger.Errorw("Failed to start transaction", "error", err)
		return err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	for _, p := range order.Products {
		_, err = tx.ExecContext(ctx,
			`DELETE FROM cart WHERE user_id = $1 AND product_id = $2`,
			order.UserID, p.ID,
		)
		if err != nil {
			s.logger.Errorw("Failed to delete product from cart",
				"userID", order.UserID,
				"productID", p.ID,
				"error", err)
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		s.logger.Errorw("Failed to commit cart cleaning transaction", "error", err)
		return err
	}
	s.logger.Infow("Cart cleaned in Postgres", "userID", order.UserID)
	return nil
}
