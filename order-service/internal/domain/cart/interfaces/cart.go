package interfaces

import "context"

type SelectedItem struct {
	ProductID int64
	Quantity  int64
}

type CartClient interface {
	GetSelectedItems(ctx context.Context, userID string) ([]SelectedItem, error)
	ConfirmPurchase(ctx context.Context, userID string) error
}
