package postgres

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
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

func (s *WalletUserStore) GetBalance(ctx context.Context, userID int64) (float64, error) {
	var balance float64
	fmt.Println("Getting balance for user ID:", userID) // Debugging line
	err := s.builder.Select("balance").
		From("wallets").
		Where("user_id = ?", userID).
		RunWith(s.db).
		Scan(&balance)
	if err != nil {
		return 0, err
	}
	return balance, nil
}

func (s *WalletUserStore) UpdateBalance(ctx context.Context, userID int64, amount float64) error {
	_, err := s.builder.Update("wallets").
		Set("balance", amount).
		Where("user_id = ?", userID).
		RunWith(s.db).
		Exec()
	return err
}
