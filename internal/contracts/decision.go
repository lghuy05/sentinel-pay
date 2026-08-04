package contracts

import "time"

type Decision string

const (
	DecisionAllow Decision = "ALLOW"
	DecisionHold  Decision = "HOLD"
	DecisionBlock Decision = "BLOCK"
)

type FraudDecisionEvent struct {
	TransactionID  string    `json:"transactionId"`
	AccountID      int64     `json:"accountId"`
	Amount         float64   `json:"amount"`
	Country        string    `json:"country"`
	BlacklistHit   bool      `json:"blacklistHit"`
	RuleScore      float64   `json:"ruleScore"`
	RuleBand       string    `json:"ruleBand"`
	MLScore        float64   `json:"mlScore"`
	MLBand         string    `json:"mlBand"`
	FinalDecision  Decision  `json:"finalDecision"`
	DecisionReason string    `json:"decisionReason"`
	ModelVersion   string    `json:"modelVersion"`
	CreatedAt      time.Time `json:"createdAt"`
}
