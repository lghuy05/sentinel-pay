package contracts

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

type feedbackFixture struct {
	TransactionID string `json:"transactionId"`
	Label         int    `json:"label"`
}

func TestContractFixturesMatchJSONContracts(t *testing.T) {
	t.Parallel()

	receiverUserID := int64(202)
	accountID := int64(101)
	amountAllow := 125.0
	amountHold := 2400.0
	amountBlock := 7500.0
	countryUS := "US"
	currencyUSD := "USD"
	deviceRisk := "LOW_RISK"
	reason := "DEVICE_BLACKLIST"
	decisionHintBlock := "BLOCK"
	ruleBandLow := "LOW"
	ruleBandMedium := "MEDIUM"
	ruleBandHigh := "HIGH"
	mlBandLow := "LOW"
	mlBandMedium := "MEDIUM"
	mlBandHigh := "HIGH"
	modelVersion := "fraud-model-v1"
	ruleVersion := 3
	lastCountry := "US"
	geoDistance := 0.0
	receiverFirstSeenDays := int64(120)
	avgAmount7D := 110.5
	avgTxPerDay := 3.2
	deviceLastSeenAt := time.Date(2026, 8, 3, 6, 30, 0, 0, time.UTC)
	lastTxTime := time.Date(2026, 8, 3, 6, 55, 0, 0, time.UTC)
	timeSinceLastTx := int64(600)

	cases := []struct {
		name string
		path string
		got  any
	}{
		{
			name: "account response",
			path: "account-response.json",
			got: AccountResponse{
				UserID:         101,
				AccountCountry: "US",
				HomeCurrency:   "USD",
				CreatedAt:      time.Date(2026, 8, 3, 7, 0, 0, 0, time.UTC),
				KycLevel:       KycLevelFull,
				Status:         AccountStatusActive,
				BalanceMinor:   125000,
			},
		},
		{
			name: "transaction request",
			path: "transaction-request.json",
			got: CreateTransactionRequest{
				TransactionID:  "tx-allow-001",
				Type:           TransactionTypeP2PTransfer,
				SenderUserID:   101,
				ReceiverUserID: &receiverUserID,
				Amount:         amountAllow,
				Currency:       "USD",
				Timestamp:      time.Date(2026, 8, 3, 7, 5, 0, 0, time.UTC),
				DeviceID:       "device-1",
			},
		},
		{
			name: "decision allow",
			path: "decision-allow.json",
			got: map[string]any{
				"id":             float64(1),
				"transactionId":  "tx-allow-001",
				"accountId":      float64(101),
				"amount":         amountAllow,
				"country":        countryUS,
				"featuresJson":   "{\"amount_usd_equivalent\":125,\"is_cross_border\":false}",
				"blacklistHit":   false,
				"ruleScore":      0.12,
				"ruleBand":       ruleBandLow,
				"ruleMatches":    "[]",
				"mlScore":        0.11,
				"mlBand":         mlBandLow,
				"finalDecision":  "ALLOW",
				"decisionReason": "No significant fraud signals detected",
				"modelVersion":   modelVersion,
				"ruleVersion":    float64(ruleVersion),
				"trueLabel":      true,
				"reviewed":       true,
				"createdAt":      "2026-08-03T07:05:02Z",
			},
		},
		{
			name: "decision hold",
			path: "decision-hold.json",
			got: map[string]any{
				"id":             float64(2),
				"transactionId":  "tx-hold-001",
				"accountId":      float64(101),
				"amount":         amountHold,
				"country":        countryUS,
				"featuresJson":   "{\"amount_usd_equivalent\":2400,\"is_cross_border\":true}",
				"blacklistHit":   false,
				"ruleScore":      0.61,
				"ruleBand":       ruleBandMedium,
				"ruleMatches":    "[\"cross-border-high-amount\"]",
				"mlScore":        0.54,
				"mlBand":         mlBandMedium,
				"finalDecision":  "HOLD",
				"decisionReason": "Elevated cross-border risk requires manual review",
				"modelVersion":   modelVersion,
				"ruleVersion":    float64(ruleVersion),
				"trueLabel":      nil,
				"reviewed":       false,
				"createdAt":      "2026-08-03T07:10:02Z",
			},
		},
		{
			name: "decision block",
			path: "decision-block.json",
			got: map[string]any{
				"id":             float64(3),
				"transactionId":  "tx-block-001",
				"accountId":      float64(101),
				"amount":         amountBlock,
				"country":        countryUS,
				"featuresJson":   "{\"amount_usd_equivalent\":7500,\"is_cross_border\":true}",
				"blacklistHit":   true,
				"ruleScore":      1.42,
				"ruleBand":       ruleBandHigh,
				"ruleMatches":    "[\"absolute-high-amount\",\"night-new-device\"]",
				"mlScore":        0.93,
				"mlBand":         mlBandHigh,
				"finalDecision":  "BLOCK",
				"decisionReason": "Blacklist hit with stacked high-risk signals",
				"modelVersion":   modelVersion,
				"ruleVersion":    float64(ruleVersion),
				"trueLabel":      nil,
				"reviewed":       false,
				"createdAt":      "2026-08-03T07:15:02Z",
			},
		},
		{
			name: "feedback request",
			path: "feedback-request.json",
			got:  feedbackFixture{TransactionID: "tx-hold-001", Label: 1},
		},
		{
			name: "ml status response",
			path: "ml-status-response.json",
			got:  map[string]any{"status": "UP", "service": "ml-service", "modelVersion": modelVersion},
		},
		{
			name: "system kafka response",
			path: "system-kafka-response.json",
			got: map[string]any{
				"status": "UP",
				"topics": []string{
					"fraud.blacklist",
					"fraud.final",
					"fraud.ml",
					"fraud.rules",
					"transactions.enriched",
					"transactions.raw",
				},
			},
		},
		{
			name: "system redis response",
			path: "system-redis-response.json",
			got:  map[string]any{"status": "UP"},
		},
		{
			name: "system services response",
			path: "system-services-response.json",
			got: map[string]any{
				"blacklist-service": map[string]any{"status": "UP", "response": map[string]any{"service": "blacklist-service", "status": "UP"}},
				"feature-extractor": map[string]any{"status": "UP", "response": map[string]any{"service": "feature-extractor", "status": "UP"}},
				"ml-service":        map[string]any{"status": "UP", "response": map[string]any{"service": "ml-service", "status": "UP"}},
				"rule-engine":       map[string]any{"status": "UP", "response": map[string]any{"service": "rule-engine", "status": "UP"}},
				"transaction-ingestor": map[string]any{
					"status":   "UP",
					"response": map[string]any{"service": "transaction-ingestor", "status": "UP"},
				},
			},
		},
		{
			name: "alert record",
			path: "alert-record.json",
			got: map[string]any{
				"id":             float64(1),
				"transactionId":  "tx-block-001",
				"decision":       "BLOCK",
				"decisionReason": "Blacklist hit with stacked high-risk signals",
				"payloadJson":    "{\"transactionId\":\"tx-block-001\",\"finalDecision\":\"BLOCK\"}",
				"decidedAt":      "2026-08-03T07:15:02Z",
				"createdAt":      "2026-08-03T07:15:03Z",
			},
		},
		{
			name: "kafka transactions raw",
			path: "kafka-transactions-raw.json",
			got: TransactionReceivedEvent{
				TransactionID:  "tx-allow-001",
				Type:           TransactionTypeP2PTransfer,
				SenderUserID:   101,
				ReceiverUserID: &receiverUserID,
				Amount:         amountAllow,
				Currency:       "USD",
				DeviceID:       "device-1",
				EventTime:      time.Date(2026, 8, 3, 7, 5, 0, 0, time.UTC),
				ReceivedAt:     time.Date(2026, 8, 3, 7, 5, 1, 0, time.UTC),
			},
		},
		{
			name: "kafka transactions enriched",
			path: "kafka-transactions-enriched.json",
			got: TransactionEnrichedEvent{
				TransactionID:            "tx-allow-001",
				Type:                     TransactionTypeP2PTransfer,
				SenderUserID:             101,
				ReceiverUserID:           &receiverUserID,
				Amount:                   amountAllow,
				Currency:                 "USD",
				DeviceID:                 "device-1",
				EventTime:                time.Date(2026, 8, 3, 7, 5, 0, 0, time.UTC),
				ReceivedAt:               time.Date(2026, 8, 3, 7, 5, 1, 0, time.UTC),
				SenderAccountCountry:     "US",
				ReceiverAccountCountry:   "US",
				SenderCountry:            "US",
				ReceiverCountry:          "US",
				LastCountry:              &lastCountry,
				GeoDistanceKM:            &geoDistance,
				SenderHomeCurrency:       "USD",
				SenderBalanceMinor:       125000,
				RiskFlag:                 &deviceRisk,
				DailyAmountUtilization:   0.12,
				DailyLimitExceeded:       false,
				SenderAccountAgeDays:     180,
				ReceiverAccountAgeDays:   120,
				ReceiverFirstSeenDays:    &receiverFirstSeenDays,
				AvgAmount7D:              &avgAmount7D,
				AvgTxPerDay:              &avgTxPerDay,
				SenderTxCount24H:         7,
				SenderTotalAmountUSD24H:  840.5,
				ReceiverInboundCount24H:  4,
				UniqueMerchantsLast24H:   0,
				DeviceLastSeenAt:         &deviceLastSeenAt,
				FeatureComputedAt:        time.Date(2026, 8, 3, 7, 5, 2, 0, time.UTC),
				CrossBorder:              false,
				Overseas:                 false,
				AmountRiskTier:           AmountRiskTierLow,
				FirstTimeContact:         false,
				AmountUSDEquivalent:      amountAllow,
				SenderReceiverTxCount24H: 2,
				TxCountLast1Min:          1,
				TxCountLast1Hour:         3,
				TxAmountLast1Hour:        250,
				AmountLast1Hour:          250,
				LastTxTime:               &lastTxTime,
				TimeSinceLastTxSeconds:   &timeSinceLastTx,
				NewDevice:                false,
			},
		},
		{
			name: "kafka fraud blacklist",
			path: "kafka-fraud-blacklist.json",
			got: BlacklistCheckEvent{
				TransactionID: "tx-block-001",
				BlacklistHit:  true,
				Reason:        &reason,
				DecisionHint:  &decisionHintBlock,
				Transaction: TransactionEnrichedEvent{
					TransactionID:            "tx-block-001",
					Type:                     TransactionTypeP2PTransfer,
					SenderUserID:             101,
					ReceiverUserID:           &receiverUserID,
					Amount:                   amountBlock,
					Currency:                 "USD",
					DeviceID:                 "device-bad-1",
					EventTime:                time.Date(2026, 8, 3, 7, 15, 0, 0, time.UTC),
					ReceivedAt:               time.Date(2026, 8, 3, 7, 15, 1, 0, time.UTC),
					SenderAccountCountry:     "US",
					ReceiverAccountCountry:   "VN",
					SenderCountry:            "US",
					ReceiverCountry:          "VN",
					SenderHomeCurrency:       "USD",
					SenderBalanceMinor:       5000000,
					DailyAmountUtilization:   0.95,
					DailyLimitExceeded:       true,
					SenderAccountAgeDays:     180,
					ReceiverAccountAgeDays:   2,
					SenderTxCount24H:         12,
					SenderTotalAmountUSD24H:  9200,
					ReceiverInboundCount24H:  18,
					UniqueMerchantsLast24H:   0,
					FeatureComputedAt:        time.Date(2026, 8, 3, 7, 15, 2, 0, time.UTC),
					CrossBorder:              true,
					Overseas:                 true,
					AmountRiskTier:           AmountRiskTierCritical,
					FirstTimeContact:         true,
					AmountUSDEquivalent:      amountBlock,
					SenderReceiverTxCount24H: 1,
					TxCountLast1Min:          4,
					TxCountLast1Hour:         9,
					TxAmountLast1Hour:        15000,
					AmountLast1Hour:          15000,
					NewDevice:                true,
				},
				EvaluatedAt: time.Date(2026, 8, 3, 7, 15, 3, 0, time.UTC),
			},
		},
		{
			name: "kafka fraud rules",
			path: "kafka-fraud-rules.json",
			got: RuleEvaluationEvent{
				TransactionID:  "tx-block-001",
				SenderUserID:   101,
				ReceiverUserID: &receiverUserID,
				Amount:         amountBlock,
				Currency:       "USD",
				RuleScore:      1.42,
				RuleBand:       ruleBandHigh,
				RuleMatches:    []string{"absolute-high-amount", "night-new-device"},
				DecisionHint:   decisionHintBlock,
				RuleVersion:    &ruleVersion,
				Features: map[string]any{
					"amount_usd_equivalent": amountBlock,
					"is_cross_border":       true,
					"is_new_device":         true,
				},
				EvaluatedAt: time.Date(2026, 8, 3, 7, 15, 4, 0, time.UTC),
			},
		},
		{
			name: "kafka fraud ml",
			path: "kafka-fraud-ml.json",
			got: MLScoreEvent{
				TransactionID: "tx-block-001",
				MLScore:       0.93,
				ModelVersion:  modelVersion,
				EvaluatedAt:   time.Date(2026, 8, 3, 7, 15, 5, 0, time.UTC),
			},
		},
		{
			name: "kafka fraud final",
			path: "kafka-fraud-final.json",
			got: FraudFinalDecisionEvent{
				TransactionID:  "tx-block-001",
				AccountID:      &accountID,
				SenderUserID:   &accountID,
				ReceiverUserID: &receiverUserID,
				Amount:         &amountBlock,
				Currency:       &currencyUSD,
				Country:        &countryUS,
				FeaturesJSON:   "{\"amount_usd_equivalent\":7500,\"is_cross_border\":true}",
				BlacklistHit:   true,
				RuleScore:      ptrFloat(1.42),
				RuleBand:       &ruleBandHigh,
				RuleMatches:    []string{"absolute-high-amount", "night-new-device"},
				MLScore:        ptrFloat(0.93),
				MLBand:         &mlBandHigh,
				FinalDecision:  FraudDecisionBlock,
				DecisionReason: "Blacklist hit with stacked high-risk signals",
				ModelVersion:   &modelVersion,
				RuleVersion:    &ruleVersion,
				DecidedAt:      time.Date(2026, 8, 3, 7, 15, 6, 0, time.UTC),
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			expected := readNormalizedFixture(t, tc.path)
			got := normalizeJSONValue(t, tc.got)
			if !reflect.DeepEqual(got, expected) {
				expectedJSON, _ := json.MarshalIndent(expected, "", "  ")
				gotJSON, _ := json.MarshalIndent(got, "", "  ")
				t.Fatalf("fixture mismatch for %s\nexpected:\n%s\nactual:\n%s", tc.path, expectedJSON, gotJSON)
			}
		})
	}

	if !containsDecisionFixture("decision-allow.json", "ALLOW") ||
		!containsDecisionFixture("decision-hold.json", "HOLD") ||
		!containsDecisionFixture("decision-block.json", "BLOCK") {
		t.Fatal("expected ALLOW/HOLD/BLOCK decision fixtures are incomplete")
	}
}

func TestRequiredKafkaTopicsIncludeFixtureTopics(t *testing.T) {
	t.Parallel()

	want := []string{
		TopicTransactionsRaw,
		TopicTransactionsEnriched,
		TopicFraudBlacklist,
		TopicFraudRules,
		TopicFraudML,
		TopicFraudFinal,
	}
	if !reflect.DeepEqual(RequiredKafkaTopics, want) {
		t.Fatalf("RequiredKafkaTopics = %v, want %v", RequiredKafkaTopics, want)
	}
}

func readNormalizedFixture(t *testing.T, name string) any {
	t.Helper()

	path := filepath.Join("..", "..", "testdata", "contracts", name)
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}

	var value any
	if err := json.Unmarshal(payload, &value); err != nil {
		t.Fatalf("unmarshal fixture %s: %v", name, err)
	}
	return value
}

func normalizeJSONValue(t *testing.T, value any) any {
	t.Helper()

	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal value: %v", err)
	}

	var normalized any
	if err := json.Unmarshal(payload, &normalized); err != nil {
		t.Fatalf("normalize value: %v", err)
	}
	return normalized
}

func containsDecisionFixture(name, decision string) bool {
	path := filepath.Join("..", "..", "testdata", "contracts", name)
	payload, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var body map[string]any
	if err := json.Unmarshal(payload, &body); err != nil {
		return false
	}
	got, _ := body["finalDecision"].(string)
	return got == decision
}

func ptrFloat(v float64) *float64 {
	return &v
}
