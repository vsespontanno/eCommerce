package postgres

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

type WalletCreatorStore struct {
	db      *sql.DB
	builder sq.StatementBuilderType
}

func NewWalletCreatorStore(db *sql.DB) *WalletCreatorStore {
	return &WalletCreatorStore{
		db:      db,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (s *WalletCreatorStore) CreateWallet(ctx context.Context, userID int64) (bool, string, error) {
	const op = "postgres.SavingWallet"
	query := s.builder.Insert("wallets").
		Columns("user_id", "balance").
		Values(userID, 0)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return false, "", err
	}

	_, err = s.db.ExecContext(ctx, sqlStr, args...)
	if err != nil {
		return false, "", fmt.Errorf("%s: %w", op, err)
	}

	return true, "Wallet created successfully", nil
}
