package intefaces

import "context"

type ProductClient interface {
	GetProductPrice(ctx context.Context, productID int64) (int, error)
}
