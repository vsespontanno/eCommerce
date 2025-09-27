package models

import "errors"

var ErrProductIsNotInCart = errors.New("product is not in cart")
var ErrTooManyProductsOfOneType = errors.New("you cannot order more than 100 products of 1 type")

var ErrProductNotInWishlist = errors.New("product is not in wishlist")
var ErrProductNotSelected = errors.New("product is not selected for order")
var ErrOrderNotFound = errors.New("order not found")
var ErrOrderAlreadyProcessed = errors.New("order already processed")
var ErrInsufficientStock = errors.New("insufficient stock")
var ErrPaymentFailed = errors.New("payment failed")
var ErrReservationFailed = errors.New("reservation failed")
