package postgres

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/vsespontanno/eCommerce/wallet-service/internal/domain/wallet/entity/apperrors"
)

type baseRepo struct {
	db      *sql.DB
	builder sq.StatementBuilderType
}

func (r *baseRepo) GetBalance(ctx context.Context, userID int64) (int64, error) {
	var balance int64
	err := r.builder.Select("balance").
		From("wallets").
		Where("user_id = ?", userID).
		RunWith(r.db).
		Scan(&balance)
	if err == sql.ErrNoRows {
		return 0, apperrors.ErrNoWallet
	}
	if err != nil {
		return 0, err
	}
	return balance, nil
}
