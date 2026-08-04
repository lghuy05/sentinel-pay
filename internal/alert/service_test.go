package alert

import (
	"strings"
	"testing"
	"time"

	"github.com/lghuy05/sentinel-pay/internal/contracts"
)

func TestSerializeFinalDecisionUsesJavaCompatibleFields(t *testing.T) {
	amount := 125.0
	currency := "USD"
	senderID := int64(100)
	event := contracts.FraudFinalDecisionEvent{
		TransactionID:  "tx-1",
		SenderUserID:   &senderID,
		Amount:         &amount,
		Currency:       &currency,
		FinalDecision:  contracts.FraudDecisionHold,
		DecisionReason: "ML_GRAY",
		DecidedAt:      time.Date(2026, 8, 4, 1, 2, 3, 0, time.UTC),
	}
	payload := serialize(event)
	for _, want := range []string{`"transactionId":"tx-1"`, `"finalDecision":"HOLD"`, `"decisionReason":"ML_GRAY"`} {
		if !strings.Contains(payload, want) {
			t.Fatalf("payload %s missing %s", payload, want)
		}
	}
}

func TestTruncateLimitsLastError(t *testing.T) {
	long := ""
	for i := 0; i < 600; i++ {
		long += "x"
	}
	if got := len(truncate(long)); got != 500 {
		t.Fatalf("truncate length = %d, want 500", got)
	}
}
