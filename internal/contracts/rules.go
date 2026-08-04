package contracts

import "time"

type RuleEvaluationEvent struct {
	TransactionID  string                   `json:"transaction_id"`
	SenderUserID   int64                    `json:"senderUserId"`
	ReceiverUserID *int64                   `json:"receiverUserId"`
	MerchantID     *int64                   `json:"merchantId"`
	Amount         float64                  `json:"amount"`
	Currency       string                   `json:"currency"`
	RuleScore      float64                  `json:"rule_score"`
	RuleBand       string                   `json:"rule_band"`
	RuleMatches    []string                 `json:"matched_rules"`
	DecisionHint   string                   `json:"decision_hint"`
	RuleVersion    *int                     `json:"ruleVersion"`
	Features       map[string]any           `json:"features"`
	EvaluatedAt    time.Time                `json:"evaluatedAt"`
	Transaction    TransactionEnrichedEvent `json:"-"`
}
