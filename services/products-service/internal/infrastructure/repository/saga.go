package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	sq "github.com/Masterminds/squirrel"
	"github.com/vsespontanno/eCommerce/services/products-service/internal/domain/apperrors"
	"github.com/vsespontanno/eCommerce/services/products-service/internal/presentation/grpc/dto"
)

// TODO: убрать дублирование кода и сделать сет более правильным, чтобы вдруг не было отрицательных значений
type SagaStore struct {
	db      *sqlx.DB
	builder sq.StatementBuilderType
}

func NewSagaStore(db *sqlx.DB) *SagaStore {
	return &SagaStore{
		db:      db,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (s *SagaStore) ReserveTxn(ctx context.Context, items []*dto.ItemRequest) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if rbErr := tx.Rollback(); rbErr != nil && rbErr != sql.ErrTxDone {
			// Log rollback errors (real errors only, ignore ErrTxDone)
			fmt.Printf("failed to rollback transaction: %v\n", rbErr)
		}
	}()

	for _, it := range items {
		qb := s.builder.
			Select("productquantity", "reserved").
			From("products").
			Where(sq.Eq{"productID": it.ProductID}).
			Suffix("FOR UPDATE")

		sqlStr, args, err := qb.ToSql()
		if err != nil {
			return err
		}

		var quantity, reserved int
		row := tx.QueryRowContext(ctx, sqlStr, args...)
		if scanErr := row.Scan(&quantity, &reserved); scanErr != nil {
			if errors.Is(scanErr, sql.ErrNoRows) {
				return fmt.Errorf("product %d not found", it.ProductID)
			}
			return scanErr
		}

		available := quantity - reserved
		if available < it.Qty {
			// явная бизнес-ошибка -> вернуть, транзакция откатится
			return fmt.Errorf("%w: productID=%d requested=%d available=%d", apperrors.ErrNotEnoughStock, it.ProductID, it.Qty, available)
		}

		// Обновляем reserved
		ub := s.builder.
			Update("products").
			Set("reserved", sq.Expr("reserved + ?", it.Qty)).
			Where(sq.Eq{"productID": it.ProductID})

		updateSQL, updateArgs, err := ub.ToSql()
		if err != nil {
			return err
		}
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
	defer func() {
		if rbErr := tx.Rollback(); rbErr != nil && rbErr != sql.ErrTxDone {
			fmt.Printf("failed to rollback transaction: %v\n", rbErr)
		}
	}()

	for _, it := range items {
		qb := s.builder.
			Select("reserved").
			From("products").
			Where(sq.Eq{"productID": it.ProductID}).
			Suffix("FOR UPDATE")

		sqlStr, args, err := qb.ToSql()
		if err != nil {
			return err
		}
		var reserved int
		if scanErr := tx.QueryRowContext(ctx, sqlStr, args...).Scan(&reserved); scanErr != nil {
			return scanErr
		}

		ub := s.builder.
			Update("products").
			Set("reserved", sq.Expr("GREATEST(reserved - ?, 0)", it.Qty)).
			Where(sq.Eq{"productID": it.ProductID})
		updateSQL, updateArgs, err := ub.ToSql()
		if err != nil {
			return err
		}
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
	defer func() {
		if rbErr := tx.Rollback(); rbErr != nil && rbErr != sql.ErrTxDone {
			fmt.Printf("failed to rollback transaction: %v\n", rbErr)
		}
	}()

	for _, it := range items {
		var quantity, reserved int
		qb := s.builder.
			Select("productquantity", "reserved").
			From("products").
			Where(sq.Eq{"productID": it.ProductID}).
			Suffix("FOR UPDATE")
		sqlStr, args, err := qb.ToSql()
		if err != nil {
			return err
		}

		if scanErr := tx.QueryRowContext(ctx, sqlStr, args...).Scan(&quantity, &reserved); scanErr != nil {
			return scanErr
		}

		if quantity < it.Qty {
			return fmt.Errorf("%w: productID=%d requested=%d available=%d", apperrors.ErrNotEnoughStock, it.ProductID, it.Qty, quantity)
		}
		if reserved < it.Qty {
			return fmt.Errorf("insufficient reserved quantity: productID=%d reserved=%d requested=%d", it.ProductID, reserved, it.Qty)
		}

		ub := s.builder.
			Update("products").
			Set("productquantity", sq.Expr("productquantity - ?", it.Qty)).
			Set("reserved", sq.Expr("reserved - ?", it.Qty)).
			Where(sq.Eq{"productID": it.ProductID})
		updateSQL, updateArgs, err := ub.ToSql()
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, updateSQL, updateArgs...); err != nil {
			return err
		}
	}

	return tx.Commit()
}
