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

// TODO: убрать дублирование кода и сделать сет более правильным, чтобы вдруг не было отрицательных значений
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

func (s *SagaStore) ReserveTxn(ctx context.Context, items []*dto.ItemRequest) error {
	// Начинаем транзакцию на уровне по-умолчанию (Read Committed в Postgres), FOR UPDATE  даст нам нужную блокировку.
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Для каждого товара: SELECT quantity, reserved FOR UPDATE , проверка, UPDATE reserved
	for _, it := range items {
		// Получаем актуальные значения с блокировкой
		qb := s.builder.
			Select("productquantity", "reserved").
			From("products").
			Where(sq.Eq{"productID": it.ProductID}).
			Suffix("FOR UPDATE")

		sqlStr, args, _ := qb.ToSql()
		fmt.Println("SQL:", sqlStr)

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
			return fmt.Errorf("%w: productID=%d requested=%d available=%d", models.ErrNotEnoughStock, it.ProductID, it.Qty, available)
		}

		// Обновляем reserved
		ub := s.builder.
			Update("products").
			Set("reserved", sq.Expr("reserved + ?", it.Qty)).
			Where(sq.Eq{"productID": it.ProductID})

		updateSQL, updateArgs, _ := ub.ToSql()
		fmt.Println("UPDATE SQL:", updateSQL)
		if _, err := tx.ExecContext(ctx, updateSQL, updateArgs...); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (s *SagaStore) ReleaseTxn(ctx context.Context, items []*dto.ItemRequest) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, it := range items {
		qb := s.builder.
			Select("reserved").
			From("products").
			Where(sq.Eq{"productID": it.ProductID}).
			Suffix("FOR UPDATE")

		sqlStr, args, _ := qb.ToSql()
		var reserved int
		fmt.Println("SQL:", sqlStr)
		if err := tx.QueryRowContext(ctx, sqlStr, args...).Scan(&reserved); err != nil {
			return err
		}

		ub := s.builder.
			Update("products").
			Set("reserved", sq.Expr("GREATEST(reserved - ?, 0)", it.Qty)).
			Where(sq.Eq{"productID": it.ProductID})
		updateSQL, updateArgs, _ := ub.ToSql()
		fmt.Println("UPDATE SQL:", updateSQL)
		if _, err := tx.ExecContext(ctx, updateSQL, updateArgs...); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *SagaStore) CommitTxn(ctx context.Context, items []*dto.ItemRequest) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, it := range items {
		var quantity, reserved int
		qb := s.builder.
			Select("productquantity", "reserved").
			From("products").
			Where(sq.Eq{"productID": it.ProductID}).
			Suffix("FOR UPDATE")
		sqlStr, args, _ := qb.ToSql()

		if err := tx.QueryRowContext(ctx, sqlStr, args...).Scan(&quantity, &reserved); err != nil {
			return err
		}

		if quantity < it.Qty {
			return fmt.Errorf("%w: productID=%d requested=%d available=%d", models.ErrNotEnoughStock, it.ProductID, it.Qty, quantity)
		}

		ub := s.builder.
			Update("products").
			Set("productquantity", sq.Expr("productquantity - ?", it.Qty)).
			Set("reserved", sq.Expr("GREATEST(reserved - ?, 0)", it.Qty)).
			Where(sq.Eq{"productID": it.ProductID})
		updateSQL, updateArgs, _ := ub.ToSql()
		if _, err := tx.ExecContext(ctx, updateSQL, updateArgs...); err != nil {
			return err
		}
	}

	return tx.Commit()
}
