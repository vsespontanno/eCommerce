package postgres

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/vsespontanno/eCommerce/cart-service/internal/domain/models"
)

var ErrNoCartFound = errors.New("no cart found")

type CartStore struct {
	db      *sql.DB
	builder sq.StatementBuilderType
}

func NewCartStore(db *sql.DB) *CartStore {
	return &CartStore{
		db:      db,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
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
