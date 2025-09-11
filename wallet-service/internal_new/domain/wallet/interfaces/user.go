package interfaces

import "context"

type UserWallet interface {
	GetBalance(ctx context.Context, userID int64) (float64, error)
	UpdateBalance(ctx context.Context, userID int64, amount float64) error
}
