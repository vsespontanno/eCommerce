package postgres

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	sq "github.com/Masterminds/squirrel"
)

type CartStore struct {
	db      *sqlx.DB
	builder sq.StatementBuilderType
}

func NewCartStore(db *sqlx.DB) *CartStore {
	return &CartStore{
		db:      db,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (s *CartStore) UpsertProductToCart(ctx context.Context, userID int64, productID int64, amountForProduct int64) (int, error) {
	// Сначала пытаемся обновить
	query := `
        UPDATE cart 
        SET quantity = quantity + 1 
        WHERE user_id = $1 AND product_id = $2
        RETURNING quantity
    `
	var quantity int
	err := s.db.QueryRowContext(ctx, query, userID, productID).Scan(&quantity)

	if err == sql.ErrNoRows {
		// Если записи нет, вставляем новую
		query = `
            INSERT INTO cart (user_id, product_id, quantity, amount_for_product)
            VALUES ($1, $2, 1, $3)
            RETURNING quantity
        `
		err = s.db.QueryRowContext(ctx, query, userID, productID, amountForProduct).Scan(&quantity)
	}

	return quantity, err
}
