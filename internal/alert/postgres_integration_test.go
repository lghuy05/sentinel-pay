package alert

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/lghuy05/sentinel-pay/internal/account"
)

func TestStoreAlertAndTransferLifecycle(t *testing.T) {
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

	store := NewStore(db)
	if err := store.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, "DELETE FROM fraud_alerts WHERE transaction_id = 'itest-alert-1'"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, "DELETE FROM applied_transfer WHERE transaction_id = 'itest-alert-1'"); err != nil {
		t.Fatal(err)
	}

	record := AlertRecord{
		TransactionID:  "itest-alert-1",
		Decision:       "BLOCK",
		DecisionReason: "ITEST",
		PayloadJSON:    "{}",
		DecidedAt:      time.Date(2026, 8, 4, 3, 0, 0, 0, time.UTC),
		CreatedAt:      time.Date(2026, 8, 4, 3, 0, 1, 0, time.UTC),
	}
	if err := store.SaveAlert(ctx, record); err != nil {
		t.Fatal(err)
	}
	if _, ok, err := store.GetAlert(ctx, record.TransactionID); err != nil || !ok {
		t.Fatalf("get alert err=%v ok=%v", err, ok)
	}

	err = store.WithTx(ctx, func(tx *sql.Tx) error {
		_, err := store.SaveTransfer(ctx, tx, Transfer{
			TransactionID: record.TransactionID,
			Status:        TransferStatusFailedRetryable,
			Attempts:      1,
			PayloadJSON:   "{}",
			CreatedAt:     time.Date(2026, 8, 4, 3, 0, 2, 0, time.UTC),
			UpdatedAt:     time.Date(2026, 8, 4, 3, 0, 3, 0, time.UTC),
		})
		return err
	})
	if err != nil {
		t.Fatal(err)
	}

	transfers, err := store.ListRetryableTransfers(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, transfer := range transfers {
		if transfer.TransactionID == record.TransactionID {
			found = true
		}
	}
	if !found {
		t.Fatal("expected retryable transfer")
	}
}
