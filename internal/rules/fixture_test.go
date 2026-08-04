package rules

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/lghuy05/sentinel-pay/internal/contracts"
)

func TestRuleEvaluationMatchesKafkaFixtureShape(t *testing.T) {
	t.Parallel()

	blacklistFixture := readRuleFixture[contracts.BlacklistCheckEvent](t, "kafka-fraud-blacklist.json")
	got := NewService().Evaluate(blacklistFixture.Transaction)
	want := readRuleFixture[contracts.RuleEvaluationEvent](t, "kafka-fraud-rules.json")

	if got.TransactionID != want.TransactionID {
		t.Fatalf("transactionId = %s, want %s", got.TransactionID, want.TransactionID)
	}
	if got.RuleBand == "" {
		t.Fatal("expected non-empty rule band")
	}
	if got.DecisionHint != want.DecisionHint {
		t.Fatalf("decisionHint = %s, want %s", got.DecisionHint, want.DecisionHint)
	}
	if len(got.RuleMatches) == 0 {
		t.Fatal("expected rule matches")
	}
}

func readRuleFixture[T any](t *testing.T, name string) T {
	t.Helper()
	var value T
	payload, err := os.ReadFile(filepath.Join("..", "..", "testdata", "contracts", name))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(payload, &value); err != nil {
		t.Fatal(err)
	}
	return value
}
