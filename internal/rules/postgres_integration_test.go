package rules

import (
	"context"
	"os"
	"testing"

	"github.com/lghuy05/sentinel-pay/internal/account"
)

func TestPostgresStoreSeedsRules(t *testing.T) {
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

	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM fraud_rules").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count < len(defaultRules()) {
		t.Fatalf("rule count = %d, want at least %d", count, len(defaultRules()))
	}
}
