package alert

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/lghuy05/sentinel-pay/internal/contracts"
)

type Service struct {
	store    *Store
	accounts *AccountClient
	logger   *slog.Logger
	now      func() time.Time
}

func NewService(store *Store, accounts *AccountClient, logger *slog.Logger) *Service {
	return &Service{store: store, accounts: accounts, logger: logger, now: time.Now}
}

func (s *Service) HandleDecision(ctx context.Context, event contracts.FraudFinalDecisionEvent) error {
	if event.TransactionID == "" || event.FinalDecision == "" {
		return nil
	}
	if event.FinalDecision == contracts.FraudDecisionAllow {
		return s.handleAllowedTransfer(ctx, event)
	}
	payload := serialize(event)
	record := AlertRecord{
		TransactionID:  event.TransactionID,
		Decision:       string(event.FinalDecision),
		DecisionReason: event.DecisionReason,
		PayloadJSON:    payload,
		DecidedAt:      nonZeroTime(event.DecidedAt, s.now()).UTC(),
		CreatedAt:      s.now().UTC(),
	}
	if err := s.store.SaveAlert(ctx, record); err != nil {
		return err
	}
	if event.FinalDecision == contracts.FraudDecisionBlock || event.FinalDecision == contracts.FraudDecisionHold {
		s.logger.Warn("ALERT", slog.String("decision", string(event.FinalDecision)), slog.String("txId", event.TransactionID), slog.String("reason", event.DecisionReason))
	}
	return nil
}

func (s *Service) handleAllowedTransfer(ctx context.Context, event contracts.FraudFinalDecisionEvent) error {
	if event.TransactionID == "" {
		return nil
	}
	payload := serialize(event)
	return s.store.WithTx(ctx, func(tx *sql.Tx) error {
		transfer, ok, err := s.store.GetTransferForUpdate(ctx, tx, event.TransactionID)
		if err != nil {
			return err
		}
		now := s.now().UTC()
		if !ok {
			transfer = Transfer{
				TransactionID: event.TransactionID,
				Status:        TransferStatusProcessing,
				Attempts:      0,
				PayloadJSON:   payload,
				CreatedAt:     now,
				UpdatedAt:     now,
			}
		}
		if transfer.Status == TransferStatusApplied {
			return nil
		}
		transfer.Status = TransferStatusProcessing
		transfer.Attempts++
		transfer.LastError = nil
		if transfer.PayloadJSON == "" {
			transfer.PayloadJSON = payload
		}
		transfer.UpdatedAt = now
		transfer, err = s.store.SaveTransfer(ctx, tx, transfer)
		if err != nil {
			return err
		}
		if err := s.validateAllowedEvent(event); err != nil {
			return s.markTransferFailed(ctx, tx, transfer, err)
		}
		if err := s.applyAllowed(ctx, event); err != nil {
			return s.markTransferFailed(ctx, tx, transfer, err)
		}
		transfer.Status = TransferStatusApplied
		transfer.UpdatedAt = s.now().UTC()
		_, err = s.store.SaveTransfer(ctx, tx, transfer)
		return err
	})
}

func (s *Service) RetryTransfer(ctx context.Context, transfer Transfer) {
	if transfer.PayloadJSON == "" {
		return
	}
	var event contracts.FraudFinalDecisionEvent
	if err := json.Unmarshal([]byte(transfer.PayloadJSON), &event); err != nil {
		_ = s.store.WithTx(ctx, func(tx *sql.Tx) error {
			transfer.Attempts++
			transfer.Status = TransferStatusFailedRetryable
			message := truncate(err.Error())
			transfer.LastError = &message
			transfer.UpdatedAt = s.now().UTC()
			_, saveErr := s.store.SaveTransfer(ctx, tx, transfer)
			return saveErr
		})
		return
	}
	if event.FinalDecision != contracts.FraudDecisionAllow {
		return
	}
	if err := s.handleAllowedTransfer(ctx, event); err != nil {
		s.logger.Warn("retry transfer failed", slog.String("transactionId", transfer.TransactionID), slog.String("error", err.Error()))
	}
}

func (s *Service) RetryFailedTransfers(ctx context.Context, limit int) error {
	transfers, err := s.store.ListRetryableTransfers(ctx, limit)
	if err != nil {
		return err
	}
	for _, transfer := range transfers {
		s.RetryTransfer(ctx, transfer)
	}
	return nil
}

func (s *Service) validateAllowedEvent(event contracts.FraudFinalDecisionEvent) error {
	if event.SenderUserID == nil || event.Amount == nil || event.Currency == nil {
		return fmt.Errorf("ALLOW event missing sender/amount/currency")
	}
	return nil
}

func (s *Service) applyAllowed(ctx context.Context, event contracts.FraudFinalDecisionEvent) error {
	if event.SenderUserID == nil || event.Amount == nil || event.Currency == nil {
		return nil
	}
	minorAmount := int64(math.Trunc(*event.Amount))
	if minorAmount <= 0 {
		return nil
	}
	if err := s.accounts.Debit(ctx, *event.SenderUserID, minorAmount, *event.Currency); err != nil {
		return err
	}
	if event.ReceiverUserID != nil {
		if err := s.accounts.TopUp(ctx, *event.ReceiverUserID, minorAmount, *event.Currency); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) markTransferFailed(ctx context.Context, tx *sql.Tx, transfer Transfer, cause error) error {
	message := truncate(cause.Error())
	transfer.Status = TransferStatusFailedRetryable
	transfer.LastError = &message
	transfer.UpdatedAt = s.now().UTC()
	if _, err := s.store.SaveTransfer(ctx, tx, transfer); err != nil {
		return err
	}
	return cause
}

func serialize(value any) string {
	payload, err := json.Marshal(value)
	if err != nil {
		return "[]"
	}
	return string(payload)
}

func truncate(message string) string {
	if len(message) > 500 {
		return message[:500]
	}
	return message
}

func nonZeroTime(value time.Time, fallback time.Time) time.Time {
	if value.IsZero() {
		return fallback
	}
	return value
}
