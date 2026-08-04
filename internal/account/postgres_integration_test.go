package account

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestPostgresStoreIntegration(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}

	ctx := context.Background()
	db, err := OpenPostgres(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := NewPostgresStore(db)
	if err := store.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, "DELETE FROM accounts WHERE user_id IN (910001, 910002)"); err != nil {
		t.Fatal(err)
	}

	created, err := store.Create(ctx, Account{
		UserID:         910001,
		AccountCountry: "US",
		HomeCurrency:   "USD",
		CreatedAt:      time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC),
		KycLevel:       "FULL",
		Status:         "ACTIVE",
		BalanceMinor:   1000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.BalanceMinor != 1000 {
		t.Fatalf("balance = %d, want 1000", created.BalanceMinor)
	}

	toppedUp, err := store.TopUp(ctx, created.UserID, 250)
	if err != nil {
		t.Fatal(err)
	}
	if toppedUp.BalanceMinor != 1250 {
		t.Fatalf("topup balance = %d, want 1250", toppedUp.BalanceMinor)
	}

	debited, err := store.Debit(ctx, created.UserID, 200)
	if err != nil {
		t.Fatal(err)
	}
	if debited.BalanceMinor != 1050 {
		t.Fatalf("debit balance = %d, want 1050", debited.BalanceMinor)
	}
}
