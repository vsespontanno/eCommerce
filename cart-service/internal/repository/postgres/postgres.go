package postgres

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/vsespontanno/eCommerce/cart-service/internal/domain/models"
	"go.uber.org/zap"
)

var ErrNoCartFound = errors.New("no cart found")

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

func (s *CartStore) GetCart(ctx context.Context, userID int64) (*models.Cart, error) {
	var cart models.Cart
	query := s.builder.
		Select("id, user_id, product_id, quantity").
		From("cart").
		Where(sq.Eq{"user_id": userID}).
		RunWith(s.db)

	rows, err := query.QueryContext(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return &models.Cart{}, ErrNoCartFound
		}
		return &models.Cart{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.CartItem
		if err := rows.Scan(&item.ID, &item.UserID, &item.ProductID, &item.Quantity); err != nil {
			return &models.Cart{}, err
		}
		cart.Items = append(cart.Items, item)
	}
	return &cart, nil
}

func (c *CartStore) CleanCart(ctx context.Context, order *models.OrderEvent) error {
	tx, err := c.db.BeginTxx(ctx, nil)
	if err != nil {
		c.logger.Errorw("Failed to start transaction", "error", err)
		return err
	}
	defer tx.Rollback()

	for _, p := range order.Products {
		_, err = tx.ExecContext(ctx,
			`DELETE FROM cart_items WHERE user_id = $1 AND product_id = $2`,
			order.UserID, p.ID,
		)
		if err != nil {
			c.logger.Errorw("Failed to delete product from cart",
				"userID", order.UserID,
				"productID", p.ID,
				"error", err)
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		c.logger.Errorw("Failed to commit cart cleaning transaction", "error", err)
		return err
	}
	c.logger.Infow("Cart cleaned in Postgres", "userID", order.UserID)
	return nil
}
