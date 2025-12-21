package interfaces

import "context"

type Auth interface {
	CreateWallet(ctx context.Context, userID int64) (bool, string, error)
}
