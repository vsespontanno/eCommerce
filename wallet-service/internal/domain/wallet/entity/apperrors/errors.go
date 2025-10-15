package apperrors

import "errors"

var (
	ErrNoWallet          = errors.New("no wallet found. would u like to create ur own?")
	ErrNotAuthorized     = errors.New("not authorized")
	ErrInsufficientFunds = errors.New("insufficient funds")
)
