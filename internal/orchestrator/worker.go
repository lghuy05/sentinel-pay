package orchestrator

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
	readers []*kafka.Reader
	dlt     *kafka.Writer
	logger  *slog.Logger
}

func NewWorker(service *Service, brokers []string, logger *slog.Logger) *Worker {
	return &Worker{
		service: service,
		readers: []*kafka.Reader{
			newReader(brokers, contracts.TopicFraudBlacklist),
			newReader(brokers, contracts.TopicFraudRules),
			newReader(brokers, contracts.TopicFraudML),
		},
		dlt:    kafkautil.NewDLTWriter(brokers),
		logger: logger,
	}
}

func newReader(brokers []string, topic string) *kafka.Reader {
	return kafkautil.NewReader(kafkautil.ReaderConfig{Brokers: brokers, Topic: topic, GroupID: ServiceName})
}

func (w *Worker) Close() error {
	var firstErr error
	for _, reader := range w.readers {
		if err := reader.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if err := w.dlt.Close(); err != nil && firstErr == nil {
		firstErr = err
	}
	return firstErr
}

func (w *Worker) Run(ctx context.Context) {
	for _, reader := range w.readers {
		go w.runReader(ctx, reader)
	}
	<-ctx.Done()
}

func (w *Worker) runReader(ctx context.Context, reader *kafka.Reader) {
	for {
		message, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			w.logger.Warn("fetch fraud signal failed", slog.String("topic", reader.Config().Topic), slog.String("error", err.Error()))
			if !kafkautil.WaitAfterFetchError(ctx) {
				return
			}
			continue
		}

		if err := w.handle(ctx, reader.Config().Topic, message.Value); err != nil {
			w.logger.Warn("handle fraud signal failed", slog.String("topic", reader.Config().Topic), slog.String("error", err.Error()))
			if publishErr := kafkautil.PublishDLT(ctx, w.dlt, message, err); publishErr != nil {
				w.logger.Warn("publish fraud signal DLT failed", slog.String("topic", reader.Config().Topic), slog.String("error", publishErr.Error()))
			} else if err := kafkautil.CommitMessage(ctx, reader, message); err != nil {
				w.logger.Warn("commit fraud signal DLT failed", slog.String("topic", reader.Config().Topic), slog.String("error", err.Error()))
			}
			continue
		}
		if err := kafkautil.CommitMessage(ctx, reader, message); err != nil {
			w.logger.Warn("commit fraud signal failed", slog.String("topic", reader.Config().Topic), slog.String("error", err.Error()))
		}
	}
}

func (w *Worker) handle(ctx context.Context, topic string, payload []byte) error {
	switch topic {
	case contracts.TopicFraudBlacklist:
		var event contracts.BlacklistCheckEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			return err
		}
		return w.service.HandleBlacklist(ctx, event)
	case contracts.TopicFraudRules:
		var event contracts.RuleEvaluationEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			return err
		}
		return w.service.HandleRule(ctx, event)
	case contracts.TopicFraudML:
		var event contracts.MLScoreEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			return err
		}
		return w.service.HandleML(ctx, event)
	default:
		return nil
	}
}
