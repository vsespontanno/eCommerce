package interfaces

import "context"

type TransactionWallet interface {
	ReserveFunds(ctx context.Context, userID int64, amount float64) error
	ReleaseFunds(ctx context.Context, userID int64, amount float64) error
	CommitFunds(ctx context.Context, userID int64, amount float64) error
}
