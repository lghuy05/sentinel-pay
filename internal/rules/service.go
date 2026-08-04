package rules

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/lghuy05/sentinel-pay/internal/contracts"
)

type Service struct {
	now func() time.Time
}

func NewService() *Service {
	return &Service{now: time.Now}
}

func (s *Service) Evaluate(event contracts.TransactionEnrichedEvent) contracts.RuleEvaluationEvent {
	matched := make([]string, 0, 12)
	totalScore := 0.0
	amountVND := toVND(event.Amount, event.Currency)

	if event.TxCountLast1Min >= 7 {
		matched = append(matched, "hard_burst_attack")
		return s.buildResult(event, 1.0, "RISK", "BLOCK", matched)
	}

	if isImpossibleGeography(event) {
		matched = append(matched, "hard_impossible_geography")
		return s.buildResult(event, 1.0, "RISK", "BLOCK", matched)
	}

	if amountVND > 10_000_000 {
		totalScore += 0.20
		matched = append(matched, "amount_gt_10m")
	}
	if amountVND > 50_000_000 {
		totalScore += 0.40
		matched = append(matched, "amount_gt_50m")
	}
	if amountVND > 200_000_000 {
		totalScore += 0.50
		matched = append(matched, "amount_gt_200m")
	}
	if amountVND > 1_000_000_000 {
		totalScore += 0.70
		matched = append(matched, "amount_gt_1b")
	}
	if event.SenderAccountAgeDays < 7 && amountVND > 5_000_000 {
		totalScore += 0.35
		matched = append(matched, "new_account_high_amount")
	}

	if event.TxCountLast1Min >= 3 {
		totalScore += 0.25
		matched = append(matched, "velocity_1m_3")
	}
	if event.TxCountLast1Min >= 5 {
		totalScore += 0.45
		matched = append(matched, "velocity_1m_5")
	}
	if event.TxCountLast1Hour >= 10 {
		totalScore += 0.20
		matched = append(matched, "velocity_1h_10")
	}
	if event.TxCountLast1Hour >= 20 {
		totalScore += 0.40
		matched = append(matched, "velocity_1h_20")
	}
	if event.AmountLast1Hour > 20_000_000 {
		totalScore += 0.25
		matched = append(matched, "amount_1h_gt_20m")
	}
	if event.AmountLast1Hour > 50_000_000 {
		totalScore += 0.45
		matched = append(matched, "amount_1h_gt_50m")
	}

	overseas := isOverseas(event)
	if overseas && amountVND > 2_000_000 {
		totalScore += 0.30
		matched = append(matched, "overseas_amount_gt_2m")
	}
	if overseas && amountVND > 200_000_000 {
		totalScore += 0.60
		matched = append(matched, "overseas_amount_gt_200m")
	}
	if overseas && event.SenderAccountAgeDays < 30 {
		totalScore += 0.35
		matched = append(matched, "overseas_new_account")
	}
	if overseas && safeInt64(event.ReceiverFirstSeenDays) < 7 {
		totalScore += 0.30
		matched = append(matched, "overseas_new_receiver")
	}

	if event.NewDevice && amountVND > 5_000_000 {
		totalScore += 0.25
		matched = append(matched, "new_device_high_amount")
	}
	if event.NewDevice && event.TxCountLast1Hour >= 5 {
		totalScore += 0.35
		matched = append(matched, "new_device_velocity_1h")
	}

	if event.SenderAccountAgeDays < 3 && amountVND > 2_000_000 {
		totalScore += 0.40
		matched = append(matched, "very_new_account_amount")
	}
	if event.SenderAccountAgeDays < 7 && event.TxCountLast1Hour >= 5 {
		totalScore += 0.35
		matched = append(matched, "new_account_velocity_1h")
	}

	avgAmount7D := safeFloat64(event.AvgAmount7D)
	avgTxPerDay := safeFloat64(event.AvgTxPerDay)
	if avgAmount7D > 0 && amountVND > 5*avgAmount7D {
		totalScore += 0.30
		matched = append(matched, "amount_gt_5x_avg7d")
	}
	if avgAmount7D > 0 && amountVND > 10*avgAmount7D {
		totalScore += 0.45
		matched = append(matched, "amount_gt_10x_avg7d")
	}
	if avgTxPerDay > 0 && float64(event.TxCountLast1Hour) > 3*avgTxPerDay {
		totalScore += 0.30
		matched = append(matched, "velocity_gt_3x_avg")
	}

	if strings.EqualFold(string(event.Type), string(contracts.TransactionTypeP2PTransfer)) {
		if amountVND > 10_000_000 {
			totalScore += 0.25
			matched = append(matched, "p2p_high_amount")
		}
		if overseas {
			totalScore += 0.35
			matched = append(matched, "p2p_overseas")
		}
	}

	score := min(totalScore, 1.0)
	band := "SAFE"
	decisionHint := "ALLOW"
	if score >= 0.40 && score <= 0.85 {
		band = "GRAY"
		decisionHint = "GRAY"
	} else if score > 0.85 {
		band = "RISK"
		decisionHint = "BLOCK"
	}

	return s.buildResult(event, score, band, decisionHint, matched)
}

func (s *Service) buildResult(event contracts.TransactionEnrichedEvent, score float64, band string, decisionHint string, matched []string) contracts.RuleEvaluationEvent {
	return contracts.RuleEvaluationEvent{
		TransactionID:  event.TransactionID,
		SenderUserID:   event.SenderUserID,
		ReceiverUserID: event.ReceiverUserID,
		MerchantID:     event.MerchantID,
		Amount:         event.Amount,
		Currency:       event.Currency,
		RuleScore:      score,
		RuleBand:       band,
		RuleMatches:    matched,
		DecisionHint:   decisionHint,
		RuleVersion:    nil,
		Features:       buildFeatureMap(event),
		EvaluatedAt:    s.now().UTC(),
		Transaction:    event,
	}
}

func toVND(amount float64, currency string) float64 {
	switch strings.ToUpper(currency) {
	case "USD":
		return amount * 25000
	case "EUR":
		return amount * 27000
	case "SGD":
		return amount * 18000
	case "VND":
		return amount
	default:
		return amount
	}
}

func isOverseas(event contracts.TransactionEnrichedEvent) bool {
	if event.SenderCountry == "" || event.ReceiverCountry == "" {
		return false
	}
	return !strings.EqualFold(event.SenderCountry, event.ReceiverCountry)
}

func isImpossibleGeography(event contracts.TransactionEnrichedEvent) bool {
	if event.LastCountry == nil || event.GeoDistanceKM == nil {
		return false
	}
	return !strings.EqualFold(event.SenderCountry, *event.LastCountry) &&
		safeInt64(event.TimeSinceLastTxSeconds) < 600 &&
		*event.GeoDistanceKM > 2000
}

func safeFloat64(value *float64) float64 {
	if value == nil {
		return 0
	}
	return *value
}

func safeInt64(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}

func buildFeatureMap(event contracts.TransactionEnrichedEvent) map[string]any {
	payload, err := json.Marshal(event)
	if err != nil {
		return map[string]any{}
	}
	var features map[string]any
	if err := json.Unmarshal(payload, &features); err != nil {
		return map[string]any{}
	}
	return features
}
