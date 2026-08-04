package ingest

import (
	"context"
	"log/slog"
	"time"

	"github.com/lghuy05/sentinel-pay/internal/platform/kafkautil"
	"github.com/segmentio/kafka-go"
)

type OutboxStore interface {
	FetchDueOutbox(ctx context.Context, limit int) ([]OutboxEvent, error)
	MarkOutboxSent(ctx context.Context, id int64) error
	MarkOutboxFailed(ctx context.Context, id int64, attempts int, message string) error
}

type Relay struct {
	store  OutboxStore
	writer *kafka.Writer
	logger *slog.Logger
}

func NewRelay(store OutboxStore, brokers []string, logger *slog.Logger) *Relay {
	return &Relay{
		store:  store,
		writer: kafkautil.NewWriter(brokers, RawTopic),
		logger: logger,
	}
}

func (r *Relay) Close() error {
	return r.writer.Close()
}

func (r *Relay) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		r.relayOnce(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (r *Relay) relayOnce(ctx context.Context) {
	events, err := r.store.FetchDueOutbox(ctx, 100)
	if err != nil {
		r.logger.Warn("fetch outbox failed", slog.String("error", err.Error()))
		return
	}
	for _, event := range events {
		err := r.writer.WriteMessages(ctx, kafka.Message{
			Key:   []byte(event.AggregateID),
			Value: []byte(event.Payload),
		})
		if err != nil {
			_ = r.store.MarkOutboxFailed(ctx, event.ID, event.AttemptCount+1, err.Error())
			r.logger.Warn("publish transaction failed", slog.String("transactionId", event.AggregateID), slog.String("error", err.Error()))
			continue
		}
		if err := r.store.MarkOutboxSent(ctx, event.ID); err != nil {
			r.logger.Warn("mark outbox sent failed", slog.Int64("outboxId", event.ID), slog.String("error", err.Error()))
		}
	}
}
