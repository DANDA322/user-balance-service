package pgstore

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"github.com/DANDA322/user-balance-service/internal/models"
	"time"

	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"
)

//go:embed migrations
var migrations embed.FS

const dateTimeFmt = "2006-01-02 15:04:05"

type DB struct {
	log *logrus.Logger
	db  *sqlx.DB
	dsn string
}

func GetPGStore(ctx context.Context, log *logrus.Logger, dsn string) (*DB, error) {
	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}
	if err = db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}
	return &DB{
		log: log,
		db:  db,
		dsn: dsn,
	}, nil
}

func (db *DB) Migrate(direction migrate.MigrationDirection) error {
	conn, err := sql.Open("pgx", db.dsn)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			db.log.Errorf("err closing migrations connection")
		}
	}()
	asserDir := func() func(string) ([]string, error) {
		return func(path string) ([]string, error) {
			dirEntry, err := migrations.ReadDir(path)
			if err != nil {
				return nil, err
			}
			entries := make([]string, 0)
			for _, e := range dirEntry {
				entries = append(entries, e.Name())
			}
			return entries, nil
		}
	}()
	asset := migrate.AssetMigrationSource{
		Asset:    migrations.ReadFile,
		AssetDir: asserDir,
		Dir:      "migrations",
	}
	_, err = migrate.Exec(conn, "postgres", asset, direction)
	return err
}

func (db *DB) GetWalletBalance(ctx context.Context, ownerId int) (*models.Wallet, error) {
	query := `
	SELECT id, balance, created_at, updated_at
	FROM wallet
	WHERE owner_id = $1`
	var wallet models.Wallet
	if err := db.db.GetContext(ctx, &wallet, query, ownerId); err != nil {
		return nil, fmt.Errorf("err executing [GetWalletBalance]: %w", err)
	}
	return &wallet, nil
}

func (db *DB) GetWalletIdByOwnerId(ctx context.Context, ownerId int) (*models.Wallet, error) {
	query := `
	SELECT id
	FROM wallet
	WHERE owner_id = $1`
	var wallet models.Wallet
	if err := db.db.GetContext(ctx, &wallet, query, ownerId); err != nil {
		return nil, fmt.Errorf("err executing [GetWalletIdByOwnerId]: %w", err)
	}
	return &wallet, nil
}

func (db *DB) UpsertDepositToWallet(ctx context.Context, ownerId int, transaction models.Transaction) error {
	query := fmt.Sprintf(`
	INSERT INTO wallet (owner_id, balance, created_at, updated_at)
	VALUES ($1, $2, '%[1]s', '%[1]s')
	ON CONFLICT (owner_id) DO UPDATE SET balance = wallet.balance + excluded.balance,
										updated_at = excluded.updated_at`,
		time.Now().Format(dateTimeFmt))
	if _, err := db.db.ExecContext(ctx, query, ownerId, transaction.Amount); err != nil {
		return fmt.Errorf("err executing [UpsertDepositToWallet]: %w", err)
	}
	wallet, err := db.GetWalletIdByOwnerId(ctx, ownerId)
	if err != nil {
		return fmt.Errorf("err executing [UpsertDepositToWallet]: %w", err)
	}
	if err := db.InsertTransaction(ctx, 0, wallet.ID, transaction); err != nil {
		return fmt.Errorf("err executing [UpsertDepositToWallet]: %w", err)
	}
	return nil
}

func (db *DB) WithdrawMoneyFromWallet(ctx context.Context, ownerId int, transaction models.Transaction) error {
	query := fmt.Sprintf(`
	UPDATE wallet SET balance = balance - $1,
	updated_at = '%s'
	WHERE owner_id = $2 AND wallet.balance - $1 > 0`,
		time.Now().Format(dateTimeFmt))
	result, err := db.db.ExecContext(ctx, query, transaction.Amount, ownerId)
	if err != nil {
		return fmt.Errorf("err executing [WithdrawMoneyFromWallet]: %w", err)
	}
	if count, _ := result.RowsAffected(); count == 0 {
		return fmt.Errorf("not enough money on the balance")
	}
	wallet, err := db.GetWalletIdByOwnerId(ctx, ownerId)
	if err != nil {
		return fmt.Errorf("err executing [WithdrawMoneyFromWallet]: %w", err)
	}
	transaction.Amount = transaction.Amount * -1
	if err := db.InsertTransaction(ctx, 0, wallet.ID, transaction); err != nil {
		return fmt.Errorf("err executing [WithdrawMoneyFromWallet]: %w", err)
	}
	return nil
}

//func (db *DB) TransferMoney(ctx context.Context, accountId int, transaction models.TransferTransaction) error {
//	query
//}

func (db *DB) InsertTransaction(ctx context.Context, walletId, targetWalletId int, transaction models.Transaction) error {
	query := fmt.Sprintf(`
	INSERT INTO transaction (wallet_id, amount, target_wallet_id, comment, timestamp)
	VALUES ($1, $2, $3, $4, '%[1]s')`,
		time.Now().Format(dateTimeFmt))
	if _, err := db.db.ExecContext(ctx, query, walletId, transaction.Amount,
		targetWalletId, transaction.Comment); err != nil {
		return fmt.Errorf("err executing [InsertTransaction]: %w", err)
	}
	return nil
}
