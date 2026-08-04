package alert

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/lghuy05/sentinel-pay/internal/contracts"
	"github.com/lghuy05/sentinel-pay/internal/platform/kafkautil"
	"github.com/segmentio/kafka-go"
)

type Worker struct {
	service *Service
	reader  *kafka.Reader
	dlt     *kafka.Writer
	logger  *slog.Logger
}

type RetryWorker struct {
	service *Service
	delay   time.Duration
	logger  *slog.Logger
}

func NewWorker(service *Service, brokers []string, logger *slog.Logger) *Worker {
	return &Worker{
		service: service,
		reader:  kafkautil.NewReader(kafkautil.ReaderConfig{Brokers: brokers, Topic: InputTopic, GroupID: ServiceName}),
		dlt:     kafkautil.NewDLTWriter(brokers),
		logger:  logger,
	}
}

func (w *Worker) Close() error {
	if err := w.reader.Close(); err != nil {
		_ = w.dlt.Close()
		return err
	}
	return w.dlt.Close()
}

func (w *Worker) Run(ctx context.Context) {
	for {
		message, err := w.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			w.logger.Warn("fetch final decision failed", slog.String("error", err.Error()))
			if !kafkautil.WaitAfterFetchError(ctx) {
				return
			}
			continue
		}
		var event contracts.FraudFinalDecisionEvent
		if err := json.Unmarshal(message.Value, &event); err != nil {
			w.logger.Warn("decode final decision failed", slog.String("error", err.Error()))
			if publishErr := kafkautil.PublishDLT(ctx, w.dlt, message, err); publishErr != nil {
				w.logger.Warn("publish final decision DLT failed", slog.String("error", publishErr.Error()))
				continue
			}
			_ = kafkautil.CommitMessage(ctx, w.reader, message)
			continue
		}
		if err := w.service.HandleDecision(ctx, event); err != nil {
			w.logger.Warn("handle final decision failed", slog.String("transactionId", event.TransactionID), slog.String("error", err.Error()))
			continue
		}
		if err := kafkautil.CommitMessage(ctx, w.reader, message); err != nil {
			w.logger.Warn("commit final decision failed", slog.String("transactionId", event.TransactionID), slog.String("error", err.Error()))
		}
	}
}

func NewRetryWorker(service *Service, delay time.Duration, logger *slog.Logger) *RetryWorker {
	if delay <= 0 {
		delay = 5 * time.Second
	}
	return &RetryWorker{service: service, delay: delay, logger: logger}
}

func (w *RetryWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.delay)
	defer ticker.Stop()
	for {
		if err := w.service.RetryFailedTransfers(ctx, 100); err != nil && ctx.Err() == nil {
			w.logger.Warn("retry failed transfers failed", slog.String("error", err.Error()))
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
