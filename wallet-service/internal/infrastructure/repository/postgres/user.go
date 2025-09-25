package postgres

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/vsespontanno/eCommerce/wallet-service/internal/domain/wallet/entity/apperrors"
)

type WalletUserStore struct {
	db      *sql.DB
	builder sq.StatementBuilderType
}

func NewWalletUserStore(db *sql.DB) *WalletUserStore {
	return &WalletUserStore{
		db:      db,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (s *WalletUserStore) GetBalance(ctx context.Context, userID int64) (int64, error) {
	var balance int64
	fmt.Println("Getting balance for user ID:", userID) // Debugging line
	err := s.builder.Select("balance").
		From("wallets").
		Where("user_id = ?", userID).
		RunWith(s.db).
		Scan(&balance)
	if err == sql.ErrNoRows {
		return 0, apperrors.ErrNoWallet
	}
	if err != nil {
		return 0, err
	}
	return balance, nil
}

func (s *WalletUserStore) UpdateBalance(ctx context.Context, userID int64, amount int64) error {
	_, err := s.builder.Update("wallets").
		Set("balance", sq.Expr("balance + ?", amount)).
		Where("user_id = ?", userID).
		RunWith(s.db).
		Exec()
	if err == sql.ErrNoRows {
		return apperrors.ErrNoWallet
	}
	return err
}

func (s *WalletUserStore) CreateWallet(ctx context.Context, userID int64) (bool, string, error) {
	_, err := s.builder.Insert("wallets").
		Columns("user_id", "balance").
		Values(userID, 0).
		RunWith(s.db).
		Exec()

	if err != nil {
		return false, "", err
	}
	return true, "Wallet created successfully", nil
}
