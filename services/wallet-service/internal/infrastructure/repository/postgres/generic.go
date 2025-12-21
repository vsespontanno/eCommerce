package postgres

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/vsespontanno/eCommerce/services/wallet-service/internal/domain/wallet/entity/apperrors"
	"go.uber.org/zap"
)

type baseRepo struct {
	db      *sql.DB
	builder sq.StatementBuilderType
	logger  *zap.SugaredLogger
}

// GetBalance returns the current balance for a user
func (r *baseRepo) GetBalance(ctx context.Context, userID int64) (int64, error) {
	var balance int64
	err := r.builder.Select("balance").
		From("wallets").
		Where(sq.Eq{"user_id": userID}).
		RunWith(r.db).
		QueryRowContext(ctx).
		Scan(&balance)

	if err == sql.ErrNoRows {
		return 0, apperrors.ErrNoWallet
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get balance for user %d: %w", userID, err)
	}

	return balance, nil
}
