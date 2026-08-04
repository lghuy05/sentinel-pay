package orchestrator

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/lghuy05/sentinel-pay/internal/account"
	"github.com/lghuy05/sentinel-pay/internal/contracts"
)

func TestStoreDecisionLifecycle(t *testing.T) {
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
	if _, err := db.ExecContext(ctx, "DELETE FROM fraud_decisions WHERE transaction_id = 'itest-orch-1'"); err != nil {
		t.Fatal(err)
	}

	accountID := int64(501)
	amount := 75.0
	currency := "USD"
	country := "US"
	ruleScore := 0.2
	ruleBand := "SAFE"
	modelVersion := "itest-model"
	ruleVersion := 1
	event := contracts.FraudFinalDecisionEvent{
		TransactionID:  "itest-orch-1",
		AccountID:      &accountID,
		SenderUserID:   &accountID,
		Amount:         &amount,
		Currency:       &currency,
		Country:        &country,
		FeaturesJSON:   "{\"amount\":75}",
		BlacklistHit:   false,
		RuleScore:      &ruleScore,
		RuleBand:       &ruleBand,
		RuleMatches:    []string{"itest-rule"},
		FinalDecision:  contracts.FraudDecisionAllow,
		DecisionReason: "ITEST",
		ModelVersion:   &modelVersion,
		RuleVersion:    &ruleVersion,
		DecidedAt:      time.Date(2026, 8, 4, 2, 0, 0, 0, time.UTC),
	}
	if err := store.Save(ctx, event); err != nil {
		t.Fatal(err)
	}

	record, ok, err := store.Get(ctx, event.TransactionID)
	if err != nil {
		t.Fatal(err)
	}
	if !ok || record.TransactionID != event.TransactionID {
		t.Fatalf("record = %+v ok=%v", record, ok)
	}

	updated, err := store.SubmitFeedback(ctx, event.TransactionID, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !updated {
		t.Fatal("expected feedback update")
	}
}
