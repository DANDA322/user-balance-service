package internal

import (
	"context"
	"fmt"
	"github.com/DANDA322/user-balance-service/internal/models"
	"github.com/sirupsen/logrus"
	"time"
)

type Database interface {
	UpsertDepositToWallet(ctx context.Context, accountID int, transaction models.Transaction) error
	WithdrawMoneyFromWallet(ctx context.Context, accountID int, transaction models.Transaction) error
	GetWallet(ctx context.Context, accountID int) (*models.Wallet, error)
	TransferMoney(ctx context.Context, accountID int, transaction models.TransferTransaction) error
	GetWalletTransactions(ctx context.Context, accountID int, sortParam string) ([]models.TransactionFullInfo, error)
	GetWalletTransactionsByDate(ctx context.Context, accountID int, timestamp time.Time) ([]models.TransactionFullInfo, error)
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

func (a *App) GetBalance(ctx context.Context, accountID int, currency string) (float64, error) {
	k := 1.0
	var err error
	if currency != "" {
		k, err = a.conv.GetRate(ctx, currency)
		if err != nil || k == 0 {
			return 0, models.ErrInvalidCurrencySymbols
		}
	}
	wallet, err := a.db.GetWallet(ctx, accountID)
	if err != nil {
		return 0, fmt.Errorf("unable to get balance: %w", err)
	}
	return wallet.Balance * k, nil
}

func (a *App) AddDepositToWallet(ctx context.Context, accountID int, transaction models.Transaction) error {
	if err := a.db.UpsertDepositToWallet(ctx, accountID, transaction); err != nil {
		return fmt.Errorf("unable to upsert deposit: %w", err)
	}
	return nil
}

func (a *App) WithdrawMoneyFromWallet(ctx context.Context, accountID int, transaction models.Transaction) error {
	if err := a.db.WithdrawMoneyFromWallet(ctx, accountID, transaction); err != nil {
		return fmt.Errorf("unable to withdraw money: %w", err)
	}
	return nil
}

func (a *App) TransferMoney(ctx context.Context, accountID int, transaction models.TransferTransaction) error {
	if err := a.db.TransferMoney(ctx, accountID, transaction); err != nil {
		return fmt.Errorf("unable to transfer money: %w", err)
	}
	return nil
}

func (a *App) GetWalletTransaction(ctx context.Context, accountID int, sortParam string) ([]models.TransactionFullInfo, error) {
	transactions, err := a.db.GetWalletTransactions(ctx, accountID, sortParam)
	if err != nil {
		return nil, fmt.Errorf("unable to get transactions: %w", err)
	}
	return transactions, nil
}

func (a *App) GetWalletTransactionsByDate(ctx context.Context, accountID int, date time.Time) ([]models.TransactionFullInfo, error) {
	transactions, err := a.db.GetWalletTransactionsByDate(ctx, accountID, date)
	if err != nil {
		return nil, fmt.Errorf("unable to get transactions: %w", err)
	}
	return transactions, nil
}
