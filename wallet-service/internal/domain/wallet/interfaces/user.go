package interfaces

import "context"

type UserWallet interface {
	GetBalance(ctx context.Context, userID int64) (int64, error)
	CreateWallet(ctx context.Context, userID int64) (bool, string, error)
	UpdateBalance(ctx context.Context, userID int64, amount int64) error
}
