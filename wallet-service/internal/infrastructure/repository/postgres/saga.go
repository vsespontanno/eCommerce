package postgres

import (
	"context"
	"database/sql"

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

func (s *SagaWalletStore) ReserveMoney(ctx context.Context, userID int64, amount int64) error {
	_, err := s.builder.Update("wallets").
		Set("reserved", sq.Expr("reserved + ?", amount)).
		Where("user_id = ?", userID).
		RunWith(s.db).
		Exec()
	if err == sql.ErrNoRows {
		return apperrors.ErrNoWallet
	}
	return err
}

func (s *SagaWalletStore) CommitMoney(ctx context.Context, userID int64, amount int64) error {
	_, err := s.builder.Update("wallets").
		Set("balance", sq.Expr("balance - ?", amount)).
		Set("reserved", sq.Expr("reserved - ?", amount)).
		Where("user_id = ?", userID).
		RunWith(s.db).
		Exec()
	if err == sql.ErrNoRows {
		return apperrors.ErrNoWallet
	}
	return err
}

func (s *SagaWalletStore) ReleaseMoney(ctx context.Context, userID int64, amount int64) error {
	_, err := s.builder.Update("wallets").
		Set("reserved", sq.Expr("reserved - ?", amount)).
		Where("user_id = ?", userID).
		RunWith(s.db).
		Exec()
	if err == sql.ErrNoRows {
		return apperrors.ErrNoWallet
	}
	return err
}
