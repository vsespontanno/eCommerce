package apperrors

import "errors"

var (
	ErrNoProductFound = errors.New("no product found")
	// ErrNotEnoughStock - недостаточно товара
	ErrNotEnoughStock = errors.New("not enough stock")
	// ErrTransient - transient ошибка, можно ретраить
	ErrTransient = errors.New("transient db error")
)
