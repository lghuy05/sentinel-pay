package rules

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/lghuy05/sentinel-pay/internal/contracts"
	"github.com/lghuy05/sentinel-pay/internal/platform/kafkautil"
	"github.com/segmentio/kafka-go"
)

type Worker struct {
	service *Service
	reader  *kafka.Reader
	writer  *kafka.Writer
	dlt     *kafka.Writer
	logger  *slog.Logger
}

func NewWorker(service *Service, brokers []string, logger *slog.Logger) *Worker {
	return &Worker{
		service: service,
		reader:  kafkautil.NewReader(kafkautil.ReaderConfig{Brokers: brokers, Topic: InputTopic, GroupID: ServiceName}),
		writer:  kafkautil.NewWriter(brokers, OutputTopic),
		dlt:     kafkautil.NewDLTWriter(brokers),
		logger:  logger,
	}
}

func (w *Worker) Close() error {
	if err := w.reader.Close(); err != nil {
		_ = w.writer.Close()
		_ = w.dlt.Close()
		return err
	}
	if err := w.writer.Close(); err != nil {
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
			w.logger.Warn("fetch blacklist check failed", slog.String("error", err.Error()))
			if !kafkautil.WaitAfterFetchError(ctx) {
				return
			}
			continue
		}

		var event contracts.BlacklistCheckEvent
		if err := json.Unmarshal(message.Value, &event); err != nil {
			w.logger.Warn("decode blacklist check failed", slog.String("error", err.Error()))
			if publishErr := kafkautil.PublishDLT(ctx, w.dlt, message, err); publishErr != nil {
				w.logger.Warn("publish blacklist check DLT failed", slog.String("error", publishErr.Error()))
				continue
			}
			_ = kafkautil.CommitMessage(ctx, w.reader, message)
			continue
		}
		if event.BlacklistHit {
			if err := kafkautil.CommitMessage(ctx, w.reader, message); err != nil {
				w.logger.Warn("commit blacklist hit failed", slog.String("transactionId", event.TransactionID), slog.String("error", err.Error()))
			}
			continue
		}

		evaluation := w.service.Evaluate(event.Transaction)
		payload, err := json.Marshal(evaluation)
		if err != nil {
			w.logger.Warn("encode rule evaluation failed", slog.String("transactionId", event.TransactionID), slog.String("error", err.Error()))
			continue
		}
		if err := w.writer.WriteMessages(ctx, kafka.Message{Key: []byte(evaluation.TransactionID), Value: payload}); err != nil {
			w.logger.Warn("publish rule evaluation failed", slog.String("transactionId", event.TransactionID), slog.String("error", err.Error()))
			continue
		}
		if err := kafkautil.CommitMessage(ctx, w.reader, message); err != nil {
			w.logger.Warn("commit blacklist check failed", slog.String("transactionId", event.TransactionID), slog.String("error", err.Error()))
		}
	}
}
