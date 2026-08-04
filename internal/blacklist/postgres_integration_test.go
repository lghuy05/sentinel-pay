package blacklist

import (
	"context"
	"os"
	"testing"

	"github.com/lghuy05/sentinel-pay/internal/account"
)

func TestPostgresStoreActiveEntries(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}

	ctx := context.Background()
	db, err := account.OpenPostgres(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := NewPostgresStore(db)
	if err := store.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, "DELETE FROM blacklist_entries WHERE value = 'itest-device'"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, "INSERT INTO blacklist_entries (type, value, active, created_at) VALUES ('DEVICE_ID', 'itest-device', TRUE, NOW())"); err != nil {
		t.Fatal(err)
	}

	entries, err := store.ActiveEntries(ctx)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, entry := range entries {
		if entry.Type == "DEVICE_ID" && entry.Value == "itest-device" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected inserted blacklist entry")
	}
}
