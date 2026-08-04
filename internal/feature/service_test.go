package feature

import (
	"testing"

	"github.com/lghuy05/sentinel-pay/internal/contracts"
)

func TestAmountRiskTierMatchesJavaThresholds(t *testing.T) {
	tests := []struct {
		amount float64
		want   contracts.AmountRiskTier
	}{
		{49.99, contracts.AmountRiskTierLow},
		{50, contracts.AmountRiskTierMedium},
		{299.99, contracts.AmountRiskTierMedium},
		{300, contracts.AmountRiskTierHigh},
		{999.99, contracts.AmountRiskTierHigh},
		{1000, contracts.AmountRiskTierCritical},
	}

	for _, tt := range tests {
		if got := amountRiskTier(tt.amount); got != tt.want {
			t.Fatalf("amountRiskTier(%v) = %s, want %s", tt.amount, got, tt.want)
		}
	}
}

func TestToUSDEquivalentMatchesJavaRates(t *testing.T) {
	tests := []struct {
		amount   float64
		currency string
		want     float64
	}{
		{25000, "VND", 1},
		{10, "USD", 10},
		{10, "EUR", 11},
		{1000, "JPY", 7},
		{42, "SGD", 42},
	}

	for _, tt := range tests {
		if got := toUSDEquivalent(tt.amount, tt.currency); got != tt.want {
			t.Fatalf("toUSDEquivalent(%v, %s) = %v, want %v", tt.amount, tt.currency, got, tt.want)
		}
	}
}

func TestCountryNormalizationAndCrossBorder(t *testing.T) {
	if got := normalizeCountry(" us "); got != "US" {
		t.Fatalf("normalizeCountry = %s, want US", got)
	}
	if got := normalizeCountry(""); got != "UNKNOWN" {
		t.Fatalf("normalizeCountry empty = %s, want UNKNOWN", got)
	}
	if !isCrossBorder("US", "VN") {
		t.Fatal("US/VN should be cross-border")
	}
	if isCrossBorder("UNKNOWN", "VN") {
		t.Fatal("UNKNOWN/VN should not be cross-border")
	}
}

func TestBuildRiskFlag(t *testing.T) {
	got := buildRiskFlag(true, "SGD")
	if got == nil || *got != "RATE_LIMIT_EXCEEDED,UNKNOWN_CURRENCY" {
		t.Fatalf("risk flag = %v", got)
	}
	if buildRiskFlag(false, "USD") != nil {
		t.Fatal("expected nil risk flag for normal USD")
	}
}
