package blacklist

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
CREATE TABLE IF NOT EXISTS blacklist_entries (
	id BIGSERIAL PRIMARY KEY,
	type VARCHAR(30) NOT NULL,
	value VARCHAR(255) NOT NULL,
	reason VARCHAR(512),
	active BOOLEAN NOT NULL DEFAULT TRUE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT uk_blacklist_entries_type_value UNIQUE (type, value)
)`)
	return err
}

func (s *PostgresStore) ActiveEntries(ctx context.Context) ([]Entry, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT type, value, active FROM blacklist_entries WHERE active = TRUE`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]Entry, 0)
	for rows.Next() {
		var entry Entry
		if err := rows.Scan(&entry.Type, &entry.Value, &entry.Active); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}
