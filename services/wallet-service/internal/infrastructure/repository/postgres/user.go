package postgres

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"
	"github.com/vsespontanno/eCommerce/services/wallet-service/internal/domain/wallet/entity/apperrors"
)

type WalletUserStore struct {
	*baseRepo
}

func NewWalletUserStore(db *sql.DB) *WalletUserStore {
	return &WalletUserStore{
		baseRepo: &baseRepo{
			db:      db,
			builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
		},
	}
}

// UpdateBalance adds amount to user's wallet balance (top up)
func (s *WalletUserStore) UpdateBalance(ctx context.Context, userID int64, amount int64) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive: %d", amount)
	}

	res, err := s.builder.Update("wallets").
		Set("balance", sq.Expr("balance + ?", amount)).
		Where(sq.Eq{"user_id": userID}).
		RunWith(s.db).
		ExecContext(ctx)

	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return apperrors.ErrNoWallet
	}

	return nil
}

// CreateWallet creates a new wallet for a user
func (s *WalletUserStore) CreateWallet(ctx context.Context, userID int64) (bool, string, error) {
	_, err := s.builder.Insert("wallets").
		Columns("user_id", "balance", "reserved").
		Values(userID, 0, 0).
		RunWith(s.db).
		ExecContext(ctx)

	if err != nil {
		// Check for unique constraint violation (wallet already exists)
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" { // unique_violation
				return false, "wallet already exists", fmt.Errorf("wallet already exists for user %d", userID)
			}
		}
		return false, "", fmt.Errorf("failed to create wallet: %w", err)
	}

	return true, "wallet created successfully", nil
}
