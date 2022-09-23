package models

import "errors"

var (
	ErrNotEnoughMoney         = errors.New("not enough money on the balance")
	ErrWalletNotFound         = errors.New("wallet not found")
	ErrInvalidCurrencySymbols = errors.New("invalid currency symbols")
)
