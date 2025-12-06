package postgres

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/vsespontanno/eCommerce/wallet-service/internal/domain/wallet/entity/apperrors"
)

type SagaWalletStore struct {
	*baseRepo
}

func NewSagaWalletStore(db *sql.DB) *SagaWalletStore {
	return &SagaWalletStore{
		baseRepo: &baseRepo{
			db:      db,
			builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
		},
	}
}

// ReserveMoney reserves funds for a transaction
// It checks if user has sufficient balance before reserving
func (s *SagaWalletStore) ReserveMoney(ctx context.Context, userID int64, amount int64) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive: %d", amount)
	}

	// Use transaction to ensure atomicity
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check if wallet exists and has sufficient balance
	var balance, reserved int64
	err = s.builder.Select("balance", "reserved").
		From("wallets").
		Where(sq.Eq{"user_id": userID}).
		RunWith(tx).
		QueryRowContext(ctx).
		Scan(&balance, &reserved)

	if err == sql.ErrNoRows {
		return apperrors.ErrNoWallet
	}
	if err != nil {
		return fmt.Errorf("failed to get wallet: %w", err)
	}

	availableBalance := balance - reserved
	if availableBalance < amount {
		return apperrors.ErrInsufficientFunds
	}

	// Reserve the funds
	res, err := s.builder.Update("wallets").
		Set("reserved", sq.Expr("reserved + ?", amount)).
		Where(sq.Eq{"user_id": userID}).
		RunWith(tx).
		ExecContext(ctx)

	if err != nil {
		return fmt.Errorf("failed to reserve funds: %w", err)
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return apperrors.ErrNoWallet
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// CommitMoney commits reserved funds (final deduction)
func (s *SagaWalletStore) CommitMoney(ctx context.Context, userID int64, amount int64) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive: %d", amount)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Verify wallet state before commit
	var balance, reserved int64
	err = s.builder.Select("balance", "reserved").
		From("wallets").
		Where(sq.Eq{"user_id": userID}).
		RunWith(tx).
		QueryRowContext(ctx).
		Scan(&balance, &reserved)

	if err == sql.ErrNoRows {
		return apperrors.ErrNoWallet
	}
	if err != nil {
		return fmt.Errorf("failed to get wallet: %w", err)
	}

	if reserved < amount {
		return fmt.Errorf("insufficient reserved funds: have %d, need %d", reserved, amount)
	}
	if balance < amount {
		return apperrors.ErrInsufficientFunds
	}

	// Commit: deduct from balance and reserved
	res, err := s.builder.Update("wallets").
		Set("balance", sq.Expr("balance - ?", amount)).
		Set("reserved", sq.Expr("reserved - ?", amount)).
		Where(sq.Eq{"user_id": userID}).
		RunWith(tx).
		ExecContext(ctx)

	if err != nil {
		return fmt.Errorf("failed to commit funds: %w", err)
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return apperrors.ErrNoWallet
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ReleaseMoney releases previously reserved funds (rollback)
func (s *SagaWalletStore) ReleaseMoney(ctx context.Context, userID int64, amount int64) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive: %d", amount)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Verify reserved amount before release
	var reserved int64
	err = s.builder.Select("reserved").
		From("wallets").
		Where(sq.Eq{"user_id": userID}).
		RunWith(tx).
		QueryRowContext(ctx).
		Scan(&reserved)

	if err == sql.ErrNoRows {
		return apperrors.ErrNoWallet
	}
	if err != nil {
		return fmt.Errorf("failed to get wallet: %w", err)
	}

	if reserved < amount {
		return fmt.Errorf("insufficient reserved funds to release: have %d, trying to release %d", reserved, amount)
	}

	// Release the funds
	res, err := s.builder.Update("wallets").
		Set("reserved", sq.Expr("reserved - ?", amount)).
		Where(sq.Eq{"user_id": userID}).
		RunWith(tx).
		ExecContext(ctx)

	if err != nil {
		return fmt.Errorf("failed to release funds: %w", err)
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return apperrors.ErrNoWallet
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
