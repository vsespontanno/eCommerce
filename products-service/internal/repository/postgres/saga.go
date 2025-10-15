package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/vsespontanno/eCommerce/products-service/internal/domain/models"
	"github.com/vsespontanno/eCommerce/products-service/internal/grpc/dto"
)

type SagaStore struct {
	db      *sql.DB
	builder sq.StatementBuilderType
}

func NewSagaStore(db *sql.DB) *SagaStore {
	return &SagaStore{
		db:      db,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (s *SagaStore) ReserveTxn(ctx context.Context, items []dto.ItemRequest) error {
	// Начинаем транзакцию на уровне по-умолчанию (Read Committed в Postgres), FOR UPDATE даст нам нужную блокировку.
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Для каждого товара: SELECT quantity, reserved FOR UPDATE, проверка, UPDATE reserved
	for _, it := range items {
		// Получаем актуальные значения с блокировкой
		qb := sq.
			Select("quantity", "reserved").
			From("products").
			Where(sq.Eq{"id": it.ProductID}).
			Suffix("FOR UPDATE")
		sqlStr, args, _ := qb.ToSql()

		var quantity, reserved int
		row := tx.QueryRowContext(ctx, sqlStr, args...)
		if err := row.Scan(&quantity, &reserved); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("product %d not found", it.ProductID)
			}
			return err
		}

		available := quantity - reserved
		if available < it.Qty {
			// явная бизнес-ошибка -> вернуть, транзакция откатится
			return fmt.Errorf("%w: product_id=%d requested=%d available=%d", models.ErrNotEnoughStock, it.ProductID, it.Qty, available)
		}

		// Обновляем reserved
		ub := sq.
			Update("products").
			Set("reserved", sq.Expr("reserved + ?", it.Qty)).
			Where(sq.Eq{"id": it.ProductID})
		updateSQL, updateArgs, _ := ub.ToSql()
		if _, err := tx.ExecContext(ctx, updateSQL, updateArgs...); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
