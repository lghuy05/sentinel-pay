package contracts

import "time"

type FraudDecision string

const (
	FraudDecisionAllow FraudDecision = "ALLOW"
	FraudDecisionHold  FraudDecision = "HOLD"
	FraudDecisionBlock FraudDecision = "BLOCK"
)

type MLScoreEvent struct {
	TransactionID string    `json:"transactionId"`
	MLScore       float64   `json:"mlScore"`
	ModelVersion  string    `json:"modelVersion"`
	EvaluatedAt   time.Time `json:"evaluatedAt"`
}

type FraudFinalDecisionEvent struct {
	TransactionID  string        `json:"transactionId"`
	AccountID      *int64        `json:"accountId"`
	SenderUserID   *int64        `json:"senderUserId"`
	ReceiverUserID *int64        `json:"receiverUserId"`
	MerchantID     *int64        `json:"merchantId"`
	Amount         *float64      `json:"amount"`
	Currency       *string       `json:"currency"`
	Country        *string       `json:"country"`
	FeaturesJSON   string        `json:"featuresJson"`
	BlacklistHit   bool          `json:"blacklistHit"`
	RuleScore      *float64      `json:"ruleScore"`
	RuleBand       *string       `json:"ruleBand"`
	RuleMatches    []string      `json:"ruleMatches"`
	MLScore        *float64      `json:"mlScore"`
	MLBand         *string       `json:"mlBand"`
	FinalDecision  FraudDecision `json:"finalDecision"`
	DecisionReason string        `json:"decisionReason"`
	ModelVersion   *string       `json:"modelVersion"`
	RuleVersion    *int          `json:"ruleVersion"`
	DecidedAt      time.Time     `json:"decidedAt"`
}
