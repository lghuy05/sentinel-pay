package rules

import (
	"context"
	"database/sql"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) Migrate(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS fraud_rules (
	id BIGSERIAL PRIMARY KEY,
	name VARCHAR(120) NOT NULL UNIQUE,
	type VARCHAR(80) NOT NULL,
	threshold NUMERIC(18, 4) NOT NULL,
	weight NUMERIC(8, 4) NOT NULL,
	enabled BOOLEAN NOT NULL DEFAULT true,
	active BOOLEAN NOT NULL DEFAULT true
)`)
	if err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `
ALTER TABLE fraud_rules
	ADD COLUMN IF NOT EXISTS name VARCHAR(120),
	ADD COLUMN IF NOT EXISTS type VARCHAR(80),
	ADD COLUMN IF NOT EXISTS threshold NUMERIC(18, 4),
	ADD COLUMN IF NOT EXISTS weight NUMERIC(8, 4),
	ADD COLUMN IF NOT EXISTS enabled BOOLEAN NOT NULL DEFAULT true,
	ADD COLUMN IF NOT EXISTS active BOOLEAN NOT NULL DEFAULT true`); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `
UPDATE fraud_rules SET
	name = COALESCE(name, CONCAT('legacy-rule-', id)),
	type = COALESCE(type, 'LEGACY'),
	threshold = COALESCE(threshold, 0),
	weight = COALESCE(weight, 0)`); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `
ALTER TABLE fraud_rules
	ALTER COLUMN name SET NOT NULL,
	ALTER COLUMN type SET NOT NULL,
	ALTER COLUMN threshold SET NOT NULL,
	ALTER COLUMN weight SET NOT NULL`); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS fraud_rules_name_key ON fraud_rules (name)`); err != nil {
		return err
	}

	for _, rule := range defaultRules() {
		if _, err := s.db.ExecContext(ctx, `
INSERT INTO fraud_rules (name, type, threshold, weight, enabled, active)
VALUES ($1, $2, $3, $4, $5, $5)
ON CONFLICT (name) DO UPDATE SET
	type = EXCLUDED.type,
	threshold = EXCLUDED.threshold,
	weight = EXCLUDED.weight,
	enabled = EXCLUDED.enabled,
	active = EXCLUDED.active`,
			rule.Name,
			rule.Type,
			rule.Threshold,
			rule.Weight,
			rule.Active,
		); err != nil {
			return err
		}
	}
	return nil
}

func defaultRules() []Rule {
	return []Rule{
		{Name: "velocity-1m", Type: "VELOCITY_1M", Threshold: 5, Weight: 0.4, Active: true},
		{Name: "amount-1h", Type: "AMOUNT_1H", Threshold: 5000000, Weight: 0.3, Active: true},
		{Name: "micro-tx-burst-1m", Type: "MICRO_TX_BURST_1M", Threshold: 6, Weight: 0.6, Active: true},
		{Name: "small-amount-spread-24h", Type: "SMALL_VALUE_SPREAD_24H", Threshold: 30, Weight: 0.5, Active: true},
		{Name: "receiver-inbound-spike-24h", Type: "RECEIVER_INBOUND_SPIKE_24H", Threshold: 25, Weight: 0.4, Active: true},
		{Name: "sender-receiver-repeat-24h", Type: "SENDER_RECEIVER_REPEAT_24H", Threshold: 3, Weight: 0.45, Active: true},
		{Name: "overseas-high-amount", Type: "OVERSEAS_HIGH_AMOUNT", Threshold: 1000, Weight: 0.6, Active: true},
		{Name: "night-new-device", Type: "NIGHT_NEW_DEVICE", Threshold: 500000, Weight: 0.5, Active: true},
		{Name: "absolute-high-amount", Type: "ABSOLUTE_HIGH_AMOUNT", Threshold: 200000000, Weight: 1.2, Active: true},
		{Name: "cross-border-high-tier", Type: "CROSS_BORDER_AMOUNT_TIER", Threshold: 3, Weight: 0.7, Active: true},
		{Name: "cross-border-high-amount", Type: "CROSS_BORDER_HIGH_AMOUNT", Threshold: 3000, Weight: 0.6, Active: true},
		{Name: "new-receiver-high-amount", Type: "NEW_RECEIVER_HIGH_AMOUNT", Threshold: 1000, Weight: 0.5, Active: true},
		{Name: "first-time-receiver-high-amount", Type: "FIRST_TIME_RECEIVER_HIGH_AMOUNT", Threshold: 1000, Weight: 0.5, Active: true},
	}
}
