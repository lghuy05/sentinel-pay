package alert

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/lghuy05/sentinel-pay/internal/contracts"
)

func TestFraudFinalFixtureDecodesForAlertHandling(t *testing.T) {
	t.Parallel()

	payload, err := os.ReadFile(filepath.Join("..", "..", "testdata", "contracts", "kafka-fraud-final.json"))
	if err != nil {
		t.Fatal(err)
	}
	var event contracts.FraudFinalDecisionEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		t.Fatal(err)
	}
	if event.TransactionID == "" {
		t.Fatal("expected transaction id")
	}
	if event.FinalDecision != contracts.FraudDecisionBlock {
		t.Fatalf("decision = %s, want BLOCK", event.FinalDecision)
	}
}
