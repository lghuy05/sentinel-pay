package orchestrator

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/lghuy05/sentinel-pay/internal/contracts"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Migrate(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS fraud_decisions (
	id BIGSERIAL PRIMARY KEY,
	transaction_id VARCHAR(255) NOT NULL UNIQUE,
	account_id BIGINT,
	amount NUMERIC(19, 4),
	country VARCHAR(255),
	features_json TEXT,
	blacklist_hit BOOLEAN NOT NULL DEFAULT false,
	rule_score DOUBLE PRECISION,
	rule_band VARCHAR(255),
	rule_matches TEXT,
	ml_score DOUBLE PRECISION,
	ml_band VARCHAR(255),
	final_decision VARCHAR(255),
	decision_reason VARCHAR(255),
	model_version VARCHAR(255),
	rule_version INTEGER,
	true_label BOOLEAN,
	reviewed BOOLEAN NOT NULL DEFAULT false,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
ALTER TABLE fraud_decisions
	ADD COLUMN IF NOT EXISTS account_id BIGINT,
	ADD COLUMN IF NOT EXISTS amount NUMERIC(19, 4),
	ADD COLUMN IF NOT EXISTS country VARCHAR(255),
	ADD COLUMN IF NOT EXISTS features_json TEXT,
	ADD COLUMN IF NOT EXISTS blacklist_hit BOOLEAN NOT NULL DEFAULT false,
	ADD COLUMN IF NOT EXISTS rule_score DOUBLE PRECISION,
	ADD COLUMN IF NOT EXISTS rule_band VARCHAR(255),
	ADD COLUMN IF NOT EXISTS rule_matches TEXT,
	ADD COLUMN IF NOT EXISTS ml_score DOUBLE PRECISION,
	ADD COLUMN IF NOT EXISTS ml_band VARCHAR(255),
	ADD COLUMN IF NOT EXISTS final_decision VARCHAR(255),
	ADD COLUMN IF NOT EXISTS decision_reason VARCHAR(255),
	ADD COLUMN IF NOT EXISTS model_version VARCHAR(255),
	ADD COLUMN IF NOT EXISTS rule_version INTEGER,
	ADD COLUMN IF NOT EXISTS true_label BOOLEAN,
	ADD COLUMN IF NOT EXISTS reviewed BOOLEAN NOT NULL DEFAULT false,
	ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT now()`)
	if err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `
UPDATE fraud_decisions
SET reviewed = false
WHERE reviewed IS NULL`); err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
ALTER TABLE fraud_decisions
	ALTER COLUMN reviewed SET DEFAULT false,
	ALTER COLUMN blacklist_hit SET DEFAULT false,
	ALTER COLUMN created_at SET DEFAULT now()`)
	return err
}

func (s *Store) Save(ctx context.Context, event contracts.FraudFinalDecisionEvent) error {
	matches, err := json.Marshal(event.RuleMatches)
	if err != nil {
		matches = []byte("[]")
	}
	decision := string(event.FinalDecision)
	_, err = s.db.ExecContext(ctx, `
INSERT INTO fraud_decisions (
	transaction_id, account_id, amount, country, features_json, blacklist_hit,
	rule_score, rule_band, rule_matches, ml_score, ml_band, final_decision,
	decision_reason, model_version, rule_version, reviewed, created_at
) VALUES (
	$1, $2, $3, $4, $5, $6,
	$7, $8, $9, $10, $11, $12,
	$13, $14, $15, false, $16
)
ON CONFLICT (transaction_id) DO UPDATE SET
	account_id = EXCLUDED.account_id,
	amount = EXCLUDED.amount,
	country = EXCLUDED.country,
	features_json = EXCLUDED.features_json,
	blacklist_hit = EXCLUDED.blacklist_hit,
	rule_score = EXCLUDED.rule_score,
	rule_band = EXCLUDED.rule_band,
	rule_matches = EXCLUDED.rule_matches,
	ml_score = EXCLUDED.ml_score,
	ml_band = EXCLUDED.ml_band,
	final_decision = EXCLUDED.final_decision,
	decision_reason = EXCLUDED.decision_reason,
	model_version = EXCLUDED.model_version,
	rule_version = EXCLUDED.rule_version,
	reviewed = fraud_decisions.reviewed,
	created_at = EXCLUDED.created_at`,
		event.TransactionID,
		event.AccountID,
		event.Amount,
		event.Country,
		event.FeaturesJSON,
		event.BlacklistHit,
		event.RuleScore,
		event.RuleBand,
		string(matches),
		event.MLScore,
		event.MLBand,
		decision,
		event.DecisionReason,
		event.ModelVersion,
		event.RuleVersion,
		event.DecidedAt,
	)
	return err
}

func (s *Store) Get(ctx context.Context, transactionID string) (DecisionRecord, bool, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, transaction_id, account_id, amount::float8, country, features_json, blacklist_hit,
	rule_score, rule_band, rule_matches, ml_score, ml_band, final_decision,
	decision_reason, model_version, rule_version, true_label, reviewed, created_at
FROM fraud_decisions
WHERE transaction_id = $1`, transactionID)
	record, err := scanDecision(row)
	if err == sql.ErrNoRows {
		return DecisionRecord{}, false, nil
	}
	return record, err == nil, err
}

func (s *Store) List(ctx context.Context, limit int, reviewed *bool) ([]DecisionRecord, error) {
	if limit < 1 {
		limit = 1
	}
	if limit > 200 {
		limit = 200
	}
	query := `
SELECT id, transaction_id, account_id, amount::float8, country, features_json, blacklist_hit,
	rule_score, rule_band, rule_matches, ml_score, ml_band, final_decision,
	decision_reason, model_version, rule_version, true_label, reviewed, created_at
FROM fraud_decisions`
	args := []any{limit}
	if reviewed != nil && !*reviewed {
		query += " WHERE reviewed = false"
	}
	query += " ORDER BY created_at DESC LIMIT $1"
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	records := make([]DecisionRecord, 0)
	for rows.Next() {
		record, err := scanDecision(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

func (s *Store) SubmitFeedback(ctx context.Context, transactionID string, label int) (bool, error) {
	trueLabel := label == 1
	result, err := s.db.ExecContext(ctx, `
UPDATE fraud_decisions
SET true_label = $2,
	reviewed = true
WHERE transaction_id = $1`, transactionID, trueLabel)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanDecision(row rowScanner) (DecisionRecord, error) {
	var record DecisionRecord
	var createdAt time.Time
	err := row.Scan(
		&record.ID,
		&record.TransactionID,
		&record.AccountID,
		&record.Amount,
		&record.Country,
		&record.FeaturesJSON,
		&record.BlacklistHit,
		&record.RuleScore,
		&record.RuleBand,
		&record.RuleMatches,
		&record.MLScore,
		&record.MLBand,
		&record.FinalDecision,
		&record.DecisionReason,
		&record.ModelVersion,
		&record.RuleVersion,
		&record.TrueLabel,
		&record.Reviewed,
		&createdAt,
	)
	record.CreatedAt = createdAt.UTC()
	return record, err
}
