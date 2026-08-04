package orchestrator

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/lghuy05/sentinel-pay/internal/account"
	"github.com/lghuy05/sentinel-pay/internal/contracts"
	"github.com/lghuy05/sentinel-pay/internal/platform/redisutil"
	"github.com/segmentio/kafka-go"
)

func TestMultiTopicDecisionIntegration(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	redisAddr := os.Getenv("TEST_REDIS_ADDR")
	if dsn == "" || redisAddr == "" {
		t.Skip("TEST_DATABASE_URL and TEST_REDIS_ADDR must be set")
	}

	host, portText, ok := strings.Cut(redisAddr, ":")
	if !ok {
		t.Fatalf("invalid TEST_REDIS_ADDR: %s", redisAddr)
	}
	redisPort, err := strconv.Atoi(portText)
	if err != nil {
		t.Fatal(err)
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

	redisClient, err := redisutil.NewClient(ctx, redisutil.Config{Host: host, Port: redisPort})
	if err != nil {
		t.Fatal(err)
	}
	defer redisClient.Close()

	transactionID := "itest-multitopic-1"
	if _, err := db.ExecContext(ctx, "DELETE FROM fraud_decisions WHERE transaction_id = $1", transactionID); err != nil {
		t.Fatal(err)
	}
	_ = redisClient.Del(ctx, aggregateKey(transactionID), "fraud:finalized:"+transactionID).Err()

	publisher := &capturingPublisher{}
	service := NewService(redisClient, store, publisher)
	worker := &Worker{service: service}

	receiverID := int64(8802)
	transaction := contracts.TransactionEnrichedEvent{
		TransactionID:          transactionID,
		Type:                   contracts.TransactionTypeP2PTransfer,
		SenderUserID:           8801,
		ReceiverUserID:         &receiverID,
		Amount:                 125,
		Currency:               "USD",
		DeviceID:               "itest-device",
		EventTime:              time.Date(2026, 8, 4, 6, 30, 0, 0, time.UTC),
		ReceivedAt:             time.Date(2026, 8, 4, 6, 30, 1, 0, time.UTC),
		SenderAccountCountry:   "US",
		ReceiverAccountCountry: "US",
		SenderCountry:          "US",
		ReceiverCountry:        "US",
	}

	blacklistPayload, err := json.Marshal(contracts.BlacklistCheckEvent{
		TransactionID: transactionID,
		BlacklistHit:  false,
		Transaction:   transaction,
		EvaluatedAt:   time.Now().UTC(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := worker.handle(ctx, contracts.TopicFraudBlacklist, blacklistPayload); err != nil {
		t.Fatal(err)
	}

	ruleBand := "GRAY"
	ruleScore := 0.45
	rulePayload, err := json.Marshal(contracts.RuleEvaluationEvent{
		TransactionID:  transactionID,
		SenderUserID:   transaction.SenderUserID,
		ReceiverUserID: transaction.ReceiverUserID,
		Amount:         transaction.Amount,
		Currency:       transaction.Currency,
		RuleScore:      ruleScore,
		RuleBand:       ruleBand,
		DecisionHint:   "HOLD",
		RuleMatches:    []string{"itest-gray-rule"},
		Features:       featureMap(transaction),
		EvaluatedAt:    time.Now().UTC(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := worker.handle(ctx, contracts.TopicFraudRules, rulePayload); err != nil {
		t.Fatal(err)
	}

	mlPayload, err := json.Marshal(contracts.MLScoreEvent{
		TransactionID: transactionID,
		MLScore:       0.2,
		ModelVersion:  "itest-model",
		EvaluatedAt:   time.Now().UTC(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := worker.handle(ctx, contracts.TopicFraudML, mlPayload); err != nil {
		t.Fatal(err)
	}

	record, ok, err := store.Get(ctx, transactionID)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected final decision record")
	}
	if record.FinalDecision == nil || *record.FinalDecision != string(contracts.FraudDecisionAllow) {
		t.Fatalf("final decision = %v, want ALLOW", record.FinalDecision)
	}
	if len(publisher.messages) != 1 {
		t.Fatalf("published messages = %d, want 1", len(publisher.messages))
	}
	if string(publisher.messages[0].Key) != transactionID {
		t.Fatalf("message key = %s, want %s", string(publisher.messages[0].Key), transactionID)
	}
}

type capturingPublisher struct {
	messages []kafka.Message
}

func (p *capturingPublisher) WriteMessages(_ context.Context, msgs ...kafka.Message) error {
	p.messages = append(p.messages, msgs...)
	return nil
}
