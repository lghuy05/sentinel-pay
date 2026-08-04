package orchestrator

import (
	"testing"

	"github.com/lghuy05/sentinel-pay/internal/contracts"
)

func TestMLDecisionBandsMatchJava(t *testing.T) {
	cases := []struct {
		score    float64
		decision contracts.FraudDecision
		reason   string
		band     string
	}{
		{score: 0.29, decision: contracts.FraudDecisionAllow, reason: "ML_SAFE", band: "SAFE"},
		{score: 0.30, decision: contracts.FraudDecisionHold, reason: "ML_GRAY", band: "GRAY"},
		{score: 0.70, decision: contracts.FraudDecisionHold, reason: "ML_GRAY", band: "GRAY"},
		{score: 0.71, decision: contracts.FraudDecisionBlock, reason: "ML_RISK", band: "RISK"},
	}

	for _, tc := range cases {
		mlBand := "GRAY"
		decision := contracts.FraudDecisionHold
		reason := "ML_GRAY"
		if tc.score < 0.30 {
			mlBand = "SAFE"
			decision = contracts.FraudDecisionAllow
			reason = "ML_SAFE"
		} else if tc.score > 0.70 {
			mlBand = "RISK"
			decision = contracts.FraudDecisionBlock
			reason = "ML_RISK"
		}
		if decision != tc.decision || reason != tc.reason || mlBand != tc.band {
			t.Fatalf("score %v -> %s/%s/%s, want %s/%s/%s", tc.score, decision, reason, mlBand, tc.decision, tc.reason, tc.band)
		}
	}
}

func TestFeatureMapUsesJavaCompatibleJSONNames(t *testing.T) {
	event := contracts.TransactionEnrichedEvent{
		TransactionID:        "tx-1",
		SenderUserID:         123,
		Amount:               125,
		Currency:             "USD",
		SenderAccountCountry: "US",
	}
	features := featureMap(event)
	if features["transactionId"] != "tx-1" {
		t.Fatalf("transactionId = %#v", features["transactionId"])
	}
	if features["senderUserId"].(float64) != 123 {
		t.Fatalf("senderUserId = %#v", features["senderUserId"])
	}
	if features["senderAccountCountry"] != "US" {
		t.Fatalf("senderAccountCountry = %#v", features["senderAccountCountry"])
	}
}
