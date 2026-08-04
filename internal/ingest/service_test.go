package ingest

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lghuy05/sentinel-pay/internal/contracts"
)

func TestIngestP2PPersistsAndCreatesOutbox(t *testing.T) {
	store := newMemoryStore()
	service := NewService(store, true)
	service.now = func() time.Time { return time.Date(2026, 8, 3, 8, 0, 1, 0, time.UTC) }
	receiverID := int64(202)

	got, err := service.Ingest(context.Background(), contracts.CreateTransactionRequest{
		TransactionID:  "tx-1",
		Type:           contracts.TransactionTypeP2PTransfer,
		SenderUserID:   101,
		ReceiverUserID: &receiverID,
		Amount:         125,
		Currency:       "USD",
		Timestamp:      time.Date(2026, 8, 3, 8, 0, 0, 0, time.UTC),
		DeviceID:       "device-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.TransactionID != "tx-1" || got.ReceivedAt.IsZero() {
		t.Fatalf("unexpected record: %+v", got)
	}
	if len(store.outbox) != 1 {
		t.Fatalf("outbox count = %d, want 1", len(store.outbox))
	}
}

func TestIngestDuplicateReturnsExistingRecord(t *testing.T) {
	store := newMemoryStore()
	service := NewService(store, true)
	receiverID := int64(202)
	request := contracts.CreateTransactionRequest{
		TransactionID:  "tx-1",
		Type:           contracts.TransactionTypeP2PTransfer,
		SenderUserID:   101,
		ReceiverUserID: &receiverID,
		Amount:         125,
		Currency:       "USD",
		Timestamp:      time.Date(2026, 8, 3, 8, 0, 0, 0, time.UTC),
		DeviceID:       "device-1",
	}

	first, err := service.Ingest(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.Ingest(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != second.ID || len(store.outbox) != 1 {
		t.Fatalf("duplicate was not idempotent first=%+v second=%+v outbox=%d", first, second, len(store.outbox))
	}
}

func TestIngestRejectsInvalidP2P(t *testing.T) {
	service := NewService(newMemoryStore(), true)

	_, err := service.Ingest(context.Background(), contracts.CreateTransactionRequest{
		TransactionID: "tx-1",
		Type:          contracts.TransactionTypeP2PTransfer,
		SenderUserID:  101,
		Amount:        125,
		Currency:      "USD",
		Timestamp:     time.Date(2026, 8, 3, 8, 0, 0, 0, time.UTC),
		DeviceID:      "device-1",
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("err = %v, want ErrInvalidInput", err)
	}
}

func TestIngestEnforcesDailyHighValueLimit(t *testing.T) {
	store := newMemoryStore()
	store.highValueCount = dailyHighValueLimit
	service := NewService(store, true)
	receiverID := int64(202)

	_, err := service.Ingest(context.Background(), contracts.CreateTransactionRequest{
		TransactionID:  "tx-1",
		Type:           contracts.TransactionTypeP2PTransfer,
		SenderUserID:   101,
		ReceiverUserID: &receiverID,
		Amount:         501,
		Currency:       "USD",
		Timestamp:      time.Date(2026, 8, 3, 8, 0, 0, 0, time.UTC),
		DeviceID:       "device-1",
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("err = %v, want ErrInvalidInput", err)
	}
}

type memoryStore struct {
	nextID         int64
	records        map[string]TransactionRecord
	outbox         []contracts.TransactionReceivedEvent
	highValueCount int64
}

func newMemoryStore() *memoryStore {
	return &memoryStore{nextID: 1, records: map[string]TransactionRecord{}}
}

func (s *memoryStore) InsertWithOutbox(_ context.Context, record TransactionRecord, event contracts.TransactionReceivedEvent) (TransactionRecord, error) {
	if _, ok := s.records[record.TransactionID]; ok {
		return TransactionRecord{}, ErrDuplicate
	}
	record.ID = s.nextID
	s.nextID++
	s.records[record.TransactionID] = record
	s.outbox = append(s.outbox, event)
	return record, nil
}

func (s *memoryStore) GetByTransactionID(_ context.Context, transactionID string) (TransactionRecord, error) {
	record, ok := s.records[transactionID]
	if !ok {
		return TransactionRecord{}, ErrNotFound
	}
	return record, nil
}

func (s *memoryStore) ListRecent(_ context.Context, _ int) ([]TransactionRecord, error) {
	records := make([]TransactionRecord, 0, len(s.records))
	for _, record := range s.records {
		records = append(records, record)
	}
	return records, nil
}

func (s *memoryStore) CountHighValueUSD(_ context.Context, _ int64, _, _ time.Time) (int64, error) {
	return s.highValueCount, nil
}
