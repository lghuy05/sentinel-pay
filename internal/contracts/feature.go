package contracts

import "time"

type AmountRiskTier string

const (
	AmountRiskTierLow      AmountRiskTier = "LOW"
	AmountRiskTierMedium   AmountRiskTier = "MEDIUM"
	AmountRiskTierHigh     AmountRiskTier = "HIGH"
	AmountRiskTierCritical AmountRiskTier = "CRITICAL"
)

type TransactionEnrichedEvent struct {
	TransactionID            string          `json:"transactionId"`
	Type                     TransactionType `json:"type"`
	SenderUserID             int64           `json:"senderUserId"`
	ReceiverUserID           *int64          `json:"receiverUserId"`
	MerchantID               *int64          `json:"merchantId"`
	Amount                   float64         `json:"amount"`
	Currency                 string          `json:"currency"`
	DeviceID                 string          `json:"deviceId"`
	EventTime                time.Time       `json:"eventTime"`
	ReceivedAt               time.Time       `json:"receivedAt"`
	SenderAccountCountry     string          `json:"senderAccountCountry"`
	ReceiverAccountCountry   string          `json:"receiverAccountCountry"`
	SenderCountry            string          `json:"senderCountry"`
	ReceiverCountry          string          `json:"receiverCountry"`
	LastCountry              *string         `json:"lastCountry"`
	GeoDistanceKM            *float64        `json:"geoDistanceKm"`
	SenderHomeCurrency       string          `json:"senderHomeCurrency"`
	SenderBalanceMinor       int64           `json:"senderBalanceMinor"`
	RiskFlag                 *string         `json:"riskFlag"`
	DailyAmountUtilization   float64         `json:"dailyAmountUtilization"`
	DailyLimitExceeded       bool            `json:"dailyLimitExceeded"`
	SenderAccountAgeDays     int64           `json:"senderAccountAgeDays"`
	ReceiverAccountAgeDays   int64           `json:"receiverAccountAgeDays"`
	ReceiverFirstSeenDays    *int64          `json:"receiverFirstSeenDays"`
	AvgAmount7D              *float64        `json:"avgAmount7d"`
	AvgTxPerDay              *float64        `json:"avgTxPerDay"`
	SenderTxCount24H         int64           `json:"senderTxCount24h"`
	SenderTotalAmountUSD24H  float64         `json:"senderTotalAmountUsd24h"`
	ReceiverInboundCount24H  int64           `json:"receiverInboundCount24h"`
	UniqueMerchantsLast24H   int64           `json:"uniqueMerchantsLast24h"`
	DeviceLastSeenAt         *time.Time      `json:"deviceLastSeenAt"`
	FeatureComputedAt        time.Time       `json:"featureComputedAt"`
	CrossBorder              bool            `json:"is_cross_border"`
	Overseas                 bool            `json:"is_overseas"`
	AmountRiskTier           AmountRiskTier  `json:"amount_risk_tier"`
	FirstTimeContact         bool            `json:"is_first_time_receiver"`
	AmountUSDEquivalent      float64         `json:"amount_usd_equivalent"`
	SenderReceiverTxCount24H int64           `json:"sender_receiver_tx_count_24h"`
	TxCountLast1Min          int64           `json:"tx_count_1m"`
	TxCountLast1Hour         int64           `json:"tx_count_1h"`
	TxAmountLast1Hour        float64         `json:"tx_amount_1hour"`
	AmountLast1Hour          float64         `json:"amount_1h"`
	LastTxTime               *time.Time      `json:"last_tx_time"`
	TimeSinceLastTxSeconds   *int64          `json:"time_since_last_tx"`
	NewDevice                bool            `json:"is_new_device"`
}
