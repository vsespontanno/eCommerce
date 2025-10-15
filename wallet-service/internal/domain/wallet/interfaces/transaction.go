package interfaces

import "context"

type TransactionWallet interface {
	GetBalance(ctx context.Context, userID int64) (int64, error)
	ReserveMoney(ctx context.Context, userID int64, amount int64) error
	ReleaseMoney(ctx context.Context, userID int64, amount int64) error
	CommitMoney(ctx context.Context, userID int64, amount int64) error
}
