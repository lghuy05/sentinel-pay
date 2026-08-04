package account

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresStore struct {
	db *sql.DB
}

func OpenPostgres(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) Migrate(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS accounts (
	user_id BIGINT PRIMARY KEY,
	account_country VARCHAR(2) NOT NULL,
	home_currency VARCHAR(3) NOT NULL,
	created_at TIMESTAMPTZ NOT NULL,
	kyc_level VARCHAR(10) NOT NULL,
	status VARCHAR(10) NOT NULL,
	balance_minor BIGINT NOT NULL,
	version BIGINT NOT NULL DEFAULT 0
)`)
	return err
}

func (s *PostgresStore) Create(ctx context.Context, account Account) (Account, error) {
	row := s.db.QueryRowContext(ctx, `
INSERT INTO accounts (
	user_id, account_country, home_currency, created_at, kyc_level, status, balance_minor, version
) VALUES ($1, $2, $3, $4, $5, $6, $7, 0)
ON CONFLICT (user_id) DO UPDATE SET
	account_country = EXCLUDED.account_country,
	home_currency = EXCLUDED.home_currency,
	created_at = EXCLUDED.created_at,
	kyc_level = EXCLUDED.kyc_level,
	status = EXCLUDED.status,
	balance_minor = EXCLUDED.balance_minor,
	version = accounts.version + 1
RETURNING user_id, account_country, home_currency, created_at, kyc_level, status, balance_minor, version`,
		account.UserID,
		account.AccountCountry,
		account.HomeCurrency,
		account.CreatedAt,
		account.KycLevel,
		account.Status,
		account.BalanceMinor,
	)
	return scanAccount(row)
}

func (s *PostgresStore) Get(ctx context.Context, userID int64) (Account, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT user_id, account_country, home_currency, created_at, kyc_level, status, balance_minor, version
FROM accounts
WHERE user_id = $1`, userID)
	return scanAccount(row)
}

func (s *PostgresStore) List(ctx context.Context, limit, offset int) ([]Account, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT user_id, account_country, home_currency, created_at, kyc_level, status, balance_minor, version
FROM accounts
ORDER BY user_id ASC
LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := make([]Account, 0)
	for rows.Next() {
		account, err := scanAccount(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	return accounts, rows.Err()
}

func (s *PostgresStore) Update(ctx context.Context, userID int64, patch AccountPatch) (Account, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Account{}, err
	}
	defer rollbackUnlessDone(tx)

	account, err := getForUpdate(ctx, tx, userID)
	if err != nil {
		return Account{}, err
	}
	if patch.AccountCountry != nil {
		account.AccountCountry = *patch.AccountCountry
	}
	if patch.CreatedAt != nil {
		account.CreatedAt = *patch.CreatedAt
	}
	if patch.KycLevel != nil {
		account.KycLevel = *patch.KycLevel
	}
	if patch.Status != nil {
		account.Status = *patch.Status
	}
	account.Version++

	row := tx.QueryRowContext(ctx, `
UPDATE accounts
SET account_country = $2, created_at = $3, kyc_level = $4, status = $5, version = $6
WHERE user_id = $1
RETURNING user_id, account_country, home_currency, created_at, kyc_level, status, balance_minor, version`,
		account.UserID,
		account.AccountCountry,
		account.CreatedAt,
		account.KycLevel,
		account.Status,
		account.Version,
	)
	updated, err := scanAccount(row)
	if err != nil {
		return Account{}, err
	}
	if err := tx.Commit(); err != nil {
		return Account{}, err
	}
	return updated, nil
}

func (s *PostgresStore) TopUp(ctx context.Context, userID int64, amountMinor int64) (Account, error) {
	return s.adjustBalance(ctx, userID, amountMinor, false)
}

func (s *PostgresStore) Debit(ctx context.Context, userID int64, amountMinor int64) (Account, error) {
	return s.adjustBalance(ctx, userID, amountMinor, true)
}

func (s *PostgresStore) Delete(ctx context.Context, userID int64) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM accounts WHERE user_id = $1`, userID)
	if err != nil {
		return err
	}
	deleted, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if deleted == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *PostgresStore) adjustBalance(ctx context.Context, userID int64, amountMinor int64, debit bool) (Account, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Account{}, err
	}
	defer rollbackUnlessDone(tx)

	account, err := getForUpdate(ctx, tx, userID)
	if err != nil {
		return Account{}, err
	}

	nextBalance := int64(0)
	if debit {
		if amountMinor > account.BalanceMinor {
			return Account{}, ErrInsufficientFunds
		}
		if account.BalanceMinor < math.MinInt64+amountMinor {
			return Account{}, ErrBalanceOverflow
		}
		nextBalance = account.BalanceMinor - amountMinor
	} else {
		if account.BalanceMinor > math.MaxInt64-amountMinor {
			return Account{}, ErrBalanceOverflow
		}
		nextBalance = account.BalanceMinor + amountMinor
	}

	row := tx.QueryRowContext(ctx, `
UPDATE accounts
SET balance_minor = $2, version = version + 1
WHERE user_id = $1
RETURNING user_id, account_country, home_currency, created_at, kyc_level, status, balance_minor, version`,
		userID,
		nextBalance,
	)
	updated, err := scanAccount(row)
	if err != nil {
		return Account{}, err
	}
	if err := tx.Commit(); err != nil {
		return Account{}, err
	}
	return updated, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanAccount(row rowScanner) (Account, error) {
	var account Account
	err := row.Scan(
		&account.UserID,
		&account.AccountCountry,
		&account.HomeCurrency,
		&account.CreatedAt,
		&account.KycLevel,
		&account.Status,
		&account.BalanceMinor,
		&account.Version,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Account{}, ErrNotFound
	}
	if err != nil {
		return Account{}, err
	}
	return account, nil
}

func getForUpdate(ctx context.Context, tx *sql.Tx, userID int64) (Account, error) {
	row := tx.QueryRowContext(ctx, `
SELECT user_id, account_country, home_currency, created_at, kyc_level, status, balance_minor, version
FROM accounts
WHERE user_id = $1
FOR UPDATE`, userID)
	return scanAccount(row)
}

func rollbackUnlessDone(tx *sql.Tx) {
	if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		fmt.Printf("rollback failed: %v\n", err)
	}
}
