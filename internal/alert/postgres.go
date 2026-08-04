package alert

import (
	"context"
	"database/sql"
	"time"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Migrate(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS fraud_alerts (
	id BIGSERIAL PRIMARY KEY,
	transaction_id VARCHAR(255) NOT NULL UNIQUE,
	decision VARCHAR(255) NOT NULL,
	decision_reason VARCHAR(255) NOT NULL,
	payload_json TEXT,
	decided_at TIMESTAMPTZ NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS applied_transfer (
	id BIGSERIAL PRIMARY KEY,
	transaction_id VARCHAR(255) NOT NULL UNIQUE,
	status VARCHAR(32) NOT NULL,
	attempts INTEGER NOT NULL DEFAULT 0,
	last_error VARCHAR(500),
	payload_json TEXT,
	created_at TIMESTAMPTZ NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL
)`); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `
ALTER TABLE fraud_alerts
	ADD COLUMN IF NOT EXISTS decision VARCHAR(255),
	ADD COLUMN IF NOT EXISTS decision_reason VARCHAR(255),
	ADD COLUMN IF NOT EXISTS payload_json TEXT,
	ADD COLUMN IF NOT EXISTS decided_at TIMESTAMPTZ,
	ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT now()`); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `
ALTER TABLE applied_transfer
	ADD COLUMN IF NOT EXISTS status VARCHAR(32),
	ADD COLUMN IF NOT EXISTS attempts INTEGER NOT NULL DEFAULT 0,
	ADD COLUMN IF NOT EXISTS last_error VARCHAR(500),
	ADD COLUMN IF NOT EXISTS payload_json TEXT,
	ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ,
	ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ`); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `
UPDATE fraud_alerts
SET created_at = COALESCE(created_at, now()),
	decided_at = COALESCE(decided_at, now()),
	decision = COALESCE(decision, 'HOLD'),
	decision_reason = COALESCE(decision_reason, 'UNKNOWN')
WHERE created_at IS NULL OR decided_at IS NULL OR decision IS NULL OR decision_reason IS NULL`)
	return err
}

func (s *Store) SaveAlert(ctx context.Context, record AlertRecord) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO fraud_alerts (transaction_id, decision, decision_reason, payload_json, decided_at, created_at)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (transaction_id) DO UPDATE SET
	decision = EXCLUDED.decision,
	decision_reason = EXCLUDED.decision_reason,
	payload_json = EXCLUDED.payload_json,
	decided_at = EXCLUDED.decided_at`,
		record.TransactionID,
		record.Decision,
		record.DecisionReason,
		record.PayloadJSON,
		record.DecidedAt,
		record.CreatedAt,
	)
	return err
}

func (s *Store) ListAlerts(ctx context.Context, limit int) ([]AlertRecord, error) {
	if limit < 1 {
		limit = 1
	}
	if limit > 200 {
		limit = 200
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT id, transaction_id, decision, decision_reason, payload_json, decided_at, created_at
FROM fraud_alerts
ORDER BY decided_at DESC
LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	records := make([]AlertRecord, 0)
	for rows.Next() {
		record, err := scanAlert(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

func (s *Store) GetAlert(ctx context.Context, transactionID string) (AlertRecord, bool, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, transaction_id, decision, decision_reason, payload_json, decided_at, created_at
FROM fraud_alerts
WHERE transaction_id = $1`, transactionID)
	record, err := scanAlert(row)
	if err == sql.ErrNoRows {
		return AlertRecord{}, false, nil
	}
	return record, err == nil, err
}

func (s *Store) GetTransferForUpdate(ctx context.Context, tx *sql.Tx, transactionID string) (Transfer, bool, error) {
	row := tx.QueryRowContext(ctx, `
SELECT id, transaction_id, status, attempts, last_error, payload_json, created_at, updated_at
FROM applied_transfer
WHERE transaction_id = $1
FOR UPDATE`, transactionID)
	transfer, err := scanTransfer(row)
	if err == sql.ErrNoRows {
		return Transfer{}, false, nil
	}
	return transfer, err == nil, err
}

func (s *Store) ListRetryableTransfers(ctx context.Context, limit int) ([]Transfer, error) {
	if limit < 1 {
		limit = 1
	}
	if limit > 100 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT id, transaction_id, status, attempts, last_error, payload_json, created_at, updated_at
FROM applied_transfer
WHERE status = $1
ORDER BY updated_at ASC
LIMIT $2`, TransferStatusFailedRetryable, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	transfers := make([]Transfer, 0)
	for rows.Next() {
		transfer, err := scanTransfer(rows)
		if err != nil {
			return nil, err
		}
		transfers = append(transfers, transfer)
	}
	return transfers, rows.Err()
}

func (s *Store) SaveTransfer(ctx context.Context, tx *sql.Tx, transfer Transfer) (Transfer, error) {
	now := time.Now().UTC()
	if transfer.CreatedAt.IsZero() {
		transfer.CreatedAt = now
	}
	if transfer.UpdatedAt.IsZero() {
		transfer.UpdatedAt = now
	}
	row := tx.QueryRowContext(ctx, `
INSERT INTO applied_transfer (transaction_id, status, attempts, last_error, payload_json, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (transaction_id) DO UPDATE SET
	status = EXCLUDED.status,
	attempts = EXCLUDED.attempts,
	last_error = EXCLUDED.last_error,
	payload_json = COALESCE(applied_transfer.payload_json, EXCLUDED.payload_json),
	updated_at = EXCLUDED.updated_at
RETURNING id, transaction_id, status, attempts, last_error, payload_json, created_at, updated_at`,
		transfer.TransactionID,
		transfer.Status,
		transfer.Attempts,
		transfer.LastError,
		transfer.PayloadJSON,
		transfer.CreatedAt,
		transfer.UpdatedAt,
	)
	return scanTransfer(row)
}

func (s *Store) WithTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanAlert(row rowScanner) (AlertRecord, error) {
	var record AlertRecord
	err := row.Scan(
		&record.ID,
		&record.TransactionID,
		&record.Decision,
		&record.DecisionReason,
		&record.PayloadJSON,
		&record.DecidedAt,
		&record.CreatedAt,
	)
	record.DecidedAt = record.DecidedAt.UTC()
	record.CreatedAt = record.CreatedAt.UTC()
	return record, err
}

func scanTransfer(row rowScanner) (Transfer, error) {
	var transfer Transfer
	err := row.Scan(
		&transfer.ID,
		&transfer.TransactionID,
		&transfer.Status,
		&transfer.Attempts,
		&transfer.LastError,
		&transfer.PayloadJSON,
		&transfer.CreatedAt,
		&transfer.UpdatedAt,
	)
	transfer.CreatedAt = transfer.CreatedAt.UTC()
	transfer.UpdatedAt = transfer.UpdatedAt.UTC()
	return transfer, err
}
