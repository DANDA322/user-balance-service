package internal

import (
	"context"
	"fmt"
	"github.com/DANDA322/user-balance-service/internal/models"
	"github.com/sirupsen/logrus"
)

type Database interface {
	UpsertDepositToWallet(ctx context.Context, accountId int, transaction models.Transaction) error
	WithdrawMoneyFromWallet(ctx context.Context, accountId int, transaction models.Transaction) error
	GetWalletBalance(ctx context.Context, accountId int) (*models.Wallet, error)
	//TransferMoney(ctx context.Context, accountId int, transaction models.TransferTransaction) error
}

type Converter interface {
	GetRate(ctx context.Context, currency string) (float64, error)
}

type App struct {
	log  *logrus.Logger
	db   Database
	conv Converter
}

func NewApp(log *logrus.Logger, db Database, conv Converter) *App {
	return &App{
		log:  log,
		db:   db,
		conv: conv,
	}
}

func (a *App) GetBalance(ctx context.Context, accountId int, currency string) (float64, error) {
	k := 1.0
	var err error
	if currency != "" {
		k, err = a.conv.GetRate(ctx, currency)
	}
	wallet, err := a.db.GetWalletBalance(ctx, accountId)
	if err != nil {
		return 0, fmt.Errorf("unable to get balance: %w", err)
	}
	return wallet.Balance * k, nil
}

func (a *App) AddDepositToWallet(ctx context.Context, accountId int, transaction models.Transaction) error {
	if err := a.db.UpsertDepositToWallet(ctx, accountId, transaction); err != nil {
		return fmt.Errorf("unable to upsert deposit: %w", err)
	}
	return nil
}

func (a *App) WithdrawMoneyFromWallet(ctx context.Context, accountId int, transaction models.Transaction) error {
	//wallet, err := a.GetBalance(ctx, accountId)
	//if err != nil {
	//	return fmt.Errorf("unable to get balance: %w", err)
	//}
	//if wallet.Balance-transaction.Amount < 0 {
	//	return fmt.Errorf("not enough money in the wallet")
	//}
	if err := a.db.WithdrawMoneyFromWallet(ctx, accountId, transaction); err != nil {
		return fmt.Errorf("unable to withdraw money: %w", err)
	}
	return nil
}

func (a *App) TransferMoney(ctx context.Context, accountId int, transaction models.TransferTransaction) error {
	//if err := a.db.TransferMoney(ctx, accountId, transaction); err != nil {
	//	return fmt.Errorf("unable to transfer money: %w", err)
	//}
	return nil
}
