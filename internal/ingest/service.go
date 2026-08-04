package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/lghuy05/sentinel-pay/internal/contracts"
)

const (
	highValueUSDThreshold = 500
	dailyHighValueLimit   = 2
)

type Store interface {
	InsertWithOutbox(ctx context.Context, record TransactionRecord, event contracts.TransactionReceivedEvent) (TransactionRecord, error)
	GetByTransactionID(ctx context.Context, transactionID string) (TransactionRecord, error)
	ListRecent(ctx context.Context, limit int) ([]TransactionRecord, error)
	CountHighValueUSD(ctx context.Context, senderUserID int64, start, end time.Time) (int64, error)
}

type Service struct {
	store                 Store
	now                   func() time.Time
	highValueLimitEnabled bool
}

func NewService(store Store, highValueLimitEnabled bool) *Service {
	return &Service{
		store:                 store,
		now:                   time.Now,
		highValueLimitEnabled: highValueLimitEnabled,
	}
}

func (s *Service) Ingest(ctx context.Context, request contracts.CreateTransactionRequest) (TransactionRecord, error) {
	if err := s.validate(ctx, request); err != nil {
		return TransactionRecord{}, err
	}

	existing, err := s.store.GetByTransactionID(ctx, request.TransactionID)
	if err == nil {
		return existing, nil
	}
	if err != ErrNotFound {
		return TransactionRecord{}, err
	}

	receivedAt := s.now().UTC()
	record := TransactionRecord{
		TransactionID:  request.TransactionID,
		Type:           string(request.Type),
		SenderUserID:   request.SenderUserID,
		ReceiverUserID: request.ReceiverUserID,
		MerchantID:     request.MerchantID,
		Amount:         request.Amount,
		Currency:       request.Currency,
		DeviceID:       request.DeviceID,
		EventTime:      request.Timestamp,
		ReceivedAt:     receivedAt,
	}
	event := contracts.TransactionReceivedEvent{
		TransactionID:  request.TransactionID,
		Type:           request.Type,
		SenderUserID:   request.SenderUserID,
		ReceiverUserID: request.ReceiverUserID,
		MerchantID:     request.MerchantID,
		Amount:         request.Amount,
		Currency:       request.Currency,
		DeviceID:       request.DeviceID,
		EventTime:      request.Timestamp,
		ReceivedAt:     receivedAt,
	}

	created, err := s.store.InsertWithOutbox(ctx, record, event)
	if err == ErrDuplicate {
		return s.store.GetByTransactionID(ctx, request.TransactionID)
	}
	return created, err
}

func (s *Service) ListRecent(ctx context.Context, limit int) ([]TransactionRecord, error) {
	capped := max(1, min(limit, 200))
	return s.store.ListRecent(ctx, capped)
}

func (s *Service) validate(ctx context.Context, request contracts.CreateTransactionRequest) error {
	if strings.TrimSpace(request.TransactionID) == "" {
		return fmt.Errorf("%w: transactionId is required", ErrInvalidInput)
	}
	if request.Type != contracts.TransactionTypeMerchantPayment && request.Type != contracts.TransactionTypeP2PTransfer {
		return fmt.Errorf("%w: invalid type", ErrInvalidInput)
	}
	if request.SenderUserID == 0 {
		return fmt.Errorf("%w: senderUserId is required", ErrInvalidInput)
	}
	if request.Amount <= 0 {
		return fmt.Errorf("%w: amount must be positive", ErrInvalidInput)
	}
	if len(strings.TrimSpace(request.Currency)) != 3 {
		return fmt.Errorf("%w: currency must be exactly 3 characters", ErrInvalidInput)
	}
	if request.Timestamp.IsZero() {
		return fmt.Errorf("%w: timestamp is required", ErrInvalidInput)
	}
	if strings.TrimSpace(request.DeviceID) == "" {
		return fmt.Errorf("%w: deviceId is required", ErrInvalidInput)
	}
	if request.Type == contracts.TransactionTypeMerchantPayment {
		if request.MerchantID == nil {
			return fmt.Errorf("%w: merchantId is required for MERCHANT_PAYMENT", ErrInvalidInput)
		}
		if request.ReceiverUserID != nil {
			return fmt.Errorf("%w: receiverUserId must be null for MERCHANT_PAYMENT", ErrInvalidInput)
		}
	}
	if request.Type == contracts.TransactionTypeP2PTransfer {
		if request.ReceiverUserID == nil {
			return fmt.Errorf("%w: receiverUserId is required for P2P_TRANSFER", ErrInvalidInput)
		}
		if request.MerchantID != nil {
			return fmt.Errorf("%w: merchantId must be null for P2P_TRANSFER", ErrInvalidInput)
		}
	}
	return s.enforceHighValueLimit(ctx, request)
}

func (s *Service) enforceHighValueLimit(ctx context.Context, request contracts.CreateTransactionRequest) error {
	if !s.highValueLimitEnabled {
		return nil
	}
	if !strings.EqualFold(request.Currency, "USD") || request.Amount <= highValueUSDThreshold {
		return nil
	}

	day := time.Date(request.Timestamp.UTC().Year(), request.Timestamp.UTC().Month(), request.Timestamp.UTC().Day(), 0, 0, 0, 0, time.UTC)
	count, err := s.store.CountHighValueUSD(ctx, request.SenderUserID, day, day.Add(24*time.Hour))
	if err != nil {
		return err
	}
	if count >= dailyHighValueLimit {
		return fmt.Errorf("%w: Daily limit exceeded for high-value USD transactions", ErrInvalidInput)
	}
	return nil
}

func MarshalEvent(event contracts.TransactionReceivedEvent) (string, error) {
	payload, err := json.Marshal(event)
	if err != nil {
		return "", err
	}
	return string(payload), nil
}
