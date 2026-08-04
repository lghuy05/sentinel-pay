package ingest

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lghuy05/sentinel-pay/internal/contracts"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) Migrate(ctx context.Context) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS transaction_records (
			id BIGSERIAL PRIMARY KEY,
			transaction_id VARCHAR(255) NOT NULL UNIQUE,
			type VARCHAR(50) NOT NULL,
			sender_user_id BIGINT NOT NULL,
			receiver_user_id BIGINT,
			merchant_id BIGINT,
			amount NUMERIC(19,4) NOT NULL,
			currency VARCHAR(3) NOT NULL,
			device_id VARCHAR(255) NOT NULL,
			event_time TIMESTAMPTZ NOT NULL,
			received_at TIMESTAMPTZ NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS outbox_event (
			id BIGSERIAL PRIMARY KEY,
			aggregate_type VARCHAR(255) NOT NULL,
			aggregate_id VARCHAR(255) NOT NULL,
			event_type VARCHAR(255) NOT NULL,
			payload TEXT,
			status VARCHAR(50) NOT NULL,
			attempt_count INTEGER NOT NULL,
			created_at TIMESTAMPTZ NOT NULL,
			next_retry_at TIMESTAMPTZ,
			published_at TIMESTAMPTZ,
			last_error TEXT,
			CONSTRAINT uk_outbox_event UNIQUE (aggregate_type, aggregate_id, event_type)
		)`,
		`ALTER TABLE transaction_records DROP COLUMN IF EXISTS ip`,
		`ALTER TABLE outbox_event ALTER COLUMN payload TYPE TEXT`,
		`ALTER TABLE outbox_event ALTER COLUMN last_error TYPE TEXT`,
		`CREATE INDEX IF NOT EXISTS idx_transaction_records_received_at ON transaction_records (received_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_transaction_records_sender_time_amount_currency ON transaction_records (sender_user_id, event_time, amount, currency)`,
		`CREATE INDEX IF NOT EXISTS idx_outbox_event_due ON outbox_event (status, next_retry_at, created_at)`,
	}

	for _, statement := range statements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

func (s *PostgresStore) InsertWithOutbox(ctx context.Context, record TransactionRecord, event contracts.TransactionReceivedEvent) (TransactionRecord, error) {
	payload, err := MarshalEvent(event)
	if err != nil {
		return TransactionRecord{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return TransactionRecord{}, err
	}
	defer rollbackUnlessDone(tx)

	row := tx.QueryRowContext(ctx, `
INSERT INTO transaction_records (
	transaction_id, type, sender_user_id, receiver_user_id, merchant_id, amount, currency, device_id, event_time, received_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, transaction_id, type, sender_user_id, receiver_user_id, merchant_id, amount::float8, currency, device_id, event_time, received_at`,
		record.TransactionID,
		record.Type,
		record.SenderUserID,
		record.ReceiverUserID,
		record.MerchantID,
		record.Amount,
		record.Currency,
		record.DeviceID,
		record.EventTime,
		record.ReceivedAt,
	)
	created, err := scanTransaction(row)
	if isUniqueViolation(err) {
		return TransactionRecord{}, ErrDuplicate
	}
	if err != nil {
		return TransactionRecord{}, err
	}

	_, err = tx.ExecContext(ctx, `
INSERT INTO outbox_event (
	aggregate_type, aggregate_id, event_type, payload, status, attempt_count, created_at, next_retry_at
) VALUES ('TRANSACTION', $1, 'TransactionReceived', $2, 'PENDING', 0, $3, $3)
ON CONFLICT (aggregate_type, aggregate_id, event_type) DO NOTHING`,
		created.TransactionID,
		payload,
		created.ReceivedAt,
	)
	if err != nil {
		return TransactionRecord{}, err
	}

	if err := tx.Commit(); err != nil {
		return TransactionRecord{}, err
	}
	return created, nil
}

func (s *PostgresStore) GetByTransactionID(ctx context.Context, transactionID string) (TransactionRecord, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, transaction_id, type, sender_user_id, receiver_user_id, merchant_id, amount::float8, currency, device_id, event_time, received_at
FROM transaction_records
WHERE transaction_id = $1`, transactionID)
	return scanTransaction(row)
}

func (s *PostgresStore) ListRecent(ctx context.Context, limit int) ([]TransactionRecord, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, transaction_id, type, sender_user_id, receiver_user_id, merchant_id, amount::float8, currency, device_id, event_time, received_at
FROM transaction_records
ORDER BY received_at DESC
LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]TransactionRecord, 0)
	for rows.Next() {
		record, err := scanTransaction(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

func (s *PostgresStore) CountHighValueUSD(ctx context.Context, senderUserID int64, start, end time.Time) (int64, error) {
	var count int64
	err := s.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM transaction_records
WHERE sender_user_id = $1
  AND event_time >= $2
  AND event_time < $3
  AND amount >= 500
  AND UPPER(currency) = 'USD'`, senderUserID, start, end).Scan(&count)
	return count, err
}

func (s *PostgresStore) FetchDueOutbox(ctx context.Context, limit int) ([]OutboxEvent, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, aggregate_type, aggregate_id, event_type, payload, status, attempt_count, created_at, next_retry_at, published_at, last_error
FROM outbox_event
WHERE status = 'PENDING' AND next_retry_at <= NOW()
ORDER BY created_at ASC
LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]OutboxEvent, 0)
	for rows.Next() {
		var event OutboxEvent
		if err := rows.Scan(
			&event.ID,
			&event.AggregateType,
			&event.AggregateID,
			&event.EventType,
			&event.Payload,
			&event.Status,
			&event.AttemptCount,
			&event.CreatedAt,
			&event.NextRetryAt,
			&event.PublishedAt,
			&event.LastError,
		); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func (s *PostgresStore) MarkOutboxSent(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `
UPDATE outbox_event
SET status = 'SENT', published_at = NOW(), last_error = NULL, next_retry_at = NULL
WHERE id = $1`, id)
	return err
}

func (s *PostgresStore) MarkOutboxFailed(ctx context.Context, id int64, attempts int, message string) error {
	status := "PENDING"
	var nextRetry any = time.Now().UTC().Add(time.Duration(attempts*5) * time.Second)
	if attempts >= 5 {
		status = "FAILED"
		nextRetry = nil
	}
	if len(message) > 500 {
		message = message[:500]
	}
	_, err := s.db.ExecContext(ctx, `
UPDATE outbox_event
SET status = $2, attempt_count = $3, last_error = $4, next_retry_at = $5
WHERE id = $1`, id, status, attempts, message, nextRetry)
	return err
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanTransaction(row rowScanner) (TransactionRecord, error) {
	var record TransactionRecord
	if err := row.Scan(
		&record.ID,
		&record.TransactionID,
		&record.Type,
		&record.SenderUserID,
		&record.ReceiverUserID,
		&record.MerchantID,
		&record.Amount,
		&record.Currency,
		&record.DeviceID,
		&record.EventTime,
		&record.ReceivedAt,
	); errors.Is(err, sql.ErrNoRows) {
		return TransactionRecord{}, ErrNotFound
	} else if err != nil {
		return TransactionRecord{}, err
	}
	return record, nil
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "SQLSTATE 23505")
}

func rollbackUnlessDone(tx *sql.Tx) {
	if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		fmt.Printf("rollback failed: %v\n", err)
	}
}
