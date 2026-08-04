package orchestrator

import "time"

const (
	ServiceName = "fraud-orchestrator"
	DefaultPort = 8085
	OutputTopic = "fraud.final"
)

type DecisionRecord struct {
	ID             int64     `json:"id"`
	TransactionID  string    `json:"transactionId"`
	AccountID      *int64    `json:"accountId"`
	Amount         *float64  `json:"amount"`
	Country        *string   `json:"country"`
	FeaturesJSON   string    `json:"featuresJson"`
	BlacklistHit   bool      `json:"blacklistHit"`
	RuleScore      *float64  `json:"ruleScore"`
	RuleBand       *string   `json:"ruleBand"`
	RuleMatches    string    `json:"ruleMatches"`
	MLScore        *float64  `json:"mlScore"`
	MLBand         *string   `json:"mlBand"`
	FinalDecision  *string   `json:"finalDecision"`
	DecisionReason string    `json:"decisionReason"`
	ModelVersion   *string   `json:"modelVersion"`
	RuleVersion    *int      `json:"ruleVersion"`
	TrueLabel      *bool     `json:"trueLabel"`
	Reviewed       bool      `json:"reviewed"`
	CreatedAt      time.Time `json:"createdAt"`
}
