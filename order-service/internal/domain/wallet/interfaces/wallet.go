package interfaces

import "context"

type WalletClient interface {
	ReserveFunds(ctx context.Context, userID string, amount int) error
	ReleaseFunds(ctx context.Context, userID string, amount int) error
	CommitFunds(ctx context.Context, userID string, amount int) error
}
