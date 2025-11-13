package models

import "errors"

var ErrProductIsNotInCart = errors.New("product is not in cart")
var ErrTooManyProductsOfOneType = errors.New("you cannot order more than 100 products of 1 type")
var ErrNoCartFound = errors.New("no cart found")
