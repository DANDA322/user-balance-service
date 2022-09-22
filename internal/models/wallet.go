package models

import "time"

type Wallet struct {
	ID        int       `json:"id" db:"id"`
	Owner     int       `json:"ownerId" db:"owner_id"`
	Balance   float64   `json:"balance" db:"balance"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Balance struct {
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
}
