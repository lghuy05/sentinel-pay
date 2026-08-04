package ingest

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/lghuy05/sentinel-pay/internal/account"
	"github.com/lghuy05/sentinel-pay/internal/contracts"
)

func TestPostgresStoreOutboxIntegration(t *testing.T) {
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
	if _, err := db.ExecContext(ctx, "DELETE FROM outbox_event WHERE aggregate_id = 'itest-ingest-1'"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, "DELETE FROM transaction_records WHERE transaction_id = 'itest-ingest-1'"); err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 8, 4, 1, 0, 0, 0, time.UTC)
	record := TransactionRecord{
		TransactionID: "itest-ingest-1",
		Type:          string(contracts.TransactionTypeP2PTransfer),
		SenderUserID:  1001,
		Amount:        42.5,
		Currency:      "USD",
		DeviceID:      "itest-device",
		EventTime:     now,
		ReceivedAt:    now,
	}
	event := contracts.TransactionReceivedEvent{
		TransactionID: "itest-ingest-1",
		Type:          contracts.TransactionTypeP2PTransfer,
		SenderUserID:  1001,
		Amount:        42.5,
		Currency:      "USD",
		DeviceID:      "itest-device",
		EventTime:     now,
		ReceivedAt:    now,
	}

	created, err := store.InsertWithOutbox(ctx, record, event)
	if err != nil {
		t.Fatal(err)
	}
	if created.TransactionID != record.TransactionID {
		t.Fatalf("transaction_id = %s, want %s", created.TransactionID, record.TransactionID)
	}

	if _, err := store.InsertWithOutbox(ctx, record, event); err != ErrDuplicate {
		t.Fatalf("duplicate err = %v, want %v", err, ErrDuplicate)
	}

	events, err := store.FetchDueOutbox(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) == 0 {
		t.Fatal("expected outbox events")
	}

	if err := store.MarkOutboxSent(ctx, events[0].ID); err != nil {
		t.Fatal(err)
	}
}
