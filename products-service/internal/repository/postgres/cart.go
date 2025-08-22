package postgres

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
)

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

func (s *CartStore) UpsertProductToCart(ctx context.Context, userID int64, productID int64) (int, error) {
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
            INSERT INTO cart (user_id, product_id, quantity)
            VALUES ($1, $2, 1)
            RETURNING quantity
        `
		err = s.db.QueryRowContext(ctx, query, userID, productID).Scan(&quantity)
	}

	return quantity, err
}
