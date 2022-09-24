package models

import "time"

type Transaction struct {
	Amount  float64 `json:"amount"`
	Comment string  `json:"comment"`
}

type TransferTransaction struct {
	Target  int     `json:"target"`
	Amount  float64 `json:"amount"`
	Comment string  `json:"comment"`
}

type TransactionFullInfo struct {
	ID             int         `json:"id" db:"id"`
	WalletID       interface{} `json:"wallet_id" db:"wallet_id"`
	Amount         float64     `json:"amount" db:"amount"`
	TargetWalletID interface{} `json:"target_wallet_id" db:"target_wallet_id"`
	Comment        string      `json:"comment" db:"comment"`
	Timestamp      time.Time   `json:"timestamp" db:"timestamp"`
}
