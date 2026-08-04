package rules

import (
	"testing"
	"time"

	"github.com/lghuy05/sentinel-pay/internal/contracts"
)

func TestEvaluateVeryNewAccountAmountMatchesJavaBehavior(t *testing.T) {
	service := NewService()
	service.now = func() time.Time { return time.Date(2026, 8, 3, 1, 2, 3, 0, time.UTC) }

	event := contracts.TransactionEnrichedEvent{
		TransactionID:        "tx-1",
		Type:                 contracts.TransactionTypeMerchantPayment,
		SenderUserID:         1001,
		Amount:               125,
		Currency:             "USD",
		SenderCountry:        "VN",
		ReceiverCountry:      "VN",
		SenderAccountAgeDays: 1,
	}

	result := service.Evaluate(event)
	if result.RuleScore != 0.40 {
		t.Fatalf("RuleScore = %v, want 0.40", result.RuleScore)
	}
	if result.RuleBand != "GRAY" {
		t.Fatalf("RuleBand = %q, want GRAY", result.RuleBand)
	}
	if result.DecisionHint != "GRAY" {
		t.Fatalf("DecisionHint = %q, want GRAY", result.DecisionHint)
	}
	if len(result.RuleMatches) != 1 || result.RuleMatches[0] != "very_new_account_amount" {
		t.Fatalf("RuleMatches = %#v, want [very_new_account_amount]", result.RuleMatches)
	}
	if result.Features["transactionId"] != "tx-1" {
		t.Fatalf("features transactionId = %#v, want tx-1", result.Features["transactionId"])
	}
}

func TestEvaluateHardBurstAttackShortCircuits(t *testing.T) {
	service := NewService()
	event := contracts.TransactionEnrichedEvent{
		TransactionID:        "tx-hard",
		Amount:               1_000_000_000,
		Currency:             "VND",
		SenderAccountAgeDays: 0,
		TxCountLast1Min:      7,
	}

	result := service.Evaluate(event)
	if result.RuleScore != 1.0 || result.RuleBand != "RISK" || result.DecisionHint != "BLOCK" {
		t.Fatalf("result = score %v band %q hint %q, want 1.0 RISK BLOCK", result.RuleScore, result.RuleBand, result.DecisionHint)
	}
	if len(result.RuleMatches) != 1 || result.RuleMatches[0] != "hard_burst_attack" {
		t.Fatalf("RuleMatches = %#v, want [hard_burst_attack]", result.RuleMatches)
	}
}

func TestEvaluateOverseasP2PStacksAndCapsScore(t *testing.T) {
	service := NewService()
	receiverFirstSeenDays := int64(1)
	event := contracts.TransactionEnrichedEvent{
		TransactionID:          "tx-p2p",
		Type:                   contracts.TransactionTypeP2PTransfer,
		Amount:                 10_000,
		Currency:               "USD",
		SenderCountry:          "VN",
		ReceiverCountry:        "US",
		SenderAccountAgeDays:   2,
		ReceiverFirstSeenDays:  &receiverFirstSeenDays,
		NewDevice:              true,
		TxCountLast1Hour:       6,
		AmountLast1Hour:        60_000_000,
		ReceiverAccountAgeDays: 0,
	}

	result := service.Evaluate(event)
	if result.RuleScore != 1.0 {
		t.Fatalf("RuleScore = %v, want capped 1.0", result.RuleScore)
	}
	if result.RuleBand != "RISK" || result.DecisionHint != "BLOCK" {
		t.Fatalf("band/hint = %q/%q, want RISK/BLOCK", result.RuleBand, result.DecisionHint)
	}
	requireMatch(t, result.RuleMatches, "overseas_new_receiver")
	requireMatch(t, result.RuleMatches, "p2p_overseas")
}

func TestEvaluateImpossibleGeography(t *testing.T) {
	service := NewService()
	last := "US"
	distance := 3000.0
	seconds := int64(500)
	event := contracts.TransactionEnrichedEvent{
		TransactionID:           "tx-geo",
		SenderCountry:           "VN",
		LastCountry:             &last,
		GeoDistanceKM:           &distance,
		TimeSinceLastTxSeconds:  &seconds,
		Amount:                  100,
		Currency:                "USD",
		SenderAccountAgeDays:    30,
		ReceiverAccountAgeDays:  30,
		ReceiverFirstSeenDays:   nil,
		ReceiverCountry:         "VN",
		TxCountLast1Min:         1,
		TxCountLast1Hour:        1,
		SenderTxCount24H:        1,
		UniqueMerchantsLast24H:  1,
		ReceiverInboundCount24H: 1,
	}

	result := service.Evaluate(event)
	if result.RuleScore != 1.0 || result.RuleBand != "RISK" || result.DecisionHint != "BLOCK" {
		t.Fatalf("result = score %v band %q hint %q, want 1.0 RISK BLOCK", result.RuleScore, result.RuleBand, result.DecisionHint)
	}
	if len(result.RuleMatches) != 1 || result.RuleMatches[0] != "hard_impossible_geography" {
		t.Fatalf("RuleMatches = %#v, want [hard_impossible_geography]", result.RuleMatches)
	}
}

func requireMatch(t *testing.T, matches []string, want string) {
	t.Helper()
	for _, match := range matches {
		if match == want {
			return
		}
	}
	t.Fatalf("RuleMatches = %#v, missing %q", matches, want)
}
