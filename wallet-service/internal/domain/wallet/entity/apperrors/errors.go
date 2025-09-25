package apperrors

import "errors"

var ErrNoWallet = errors.New("no wallet found. would u like to create ur own?")
var ErrNotAuthorized = errors.New("not authorized")
