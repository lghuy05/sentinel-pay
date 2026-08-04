package contracts

import "time"

type BlacklistCheckEvent struct {
	TransactionID string                   `json:"transactionId"`
	BlacklistHit  bool                     `json:"blacklistHit"`
	Reason        *string                  `json:"reason"`
	DecisionHint  *string                  `json:"decisionHint"`
	Transaction   TransactionEnrichedEvent `json:"transaction"`
	EvaluatedAt   time.Time                `json:"evaluatedAt"`
}
