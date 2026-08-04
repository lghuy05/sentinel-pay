package feature

import "errors"

const (
	ServiceName = "feature-extractor"
	DefaultPort = 8082
	InputTopic  = "transactions.raw"
	OutputTopic = "transactions.enriched"
)

var ErrAccountLookup = errors.New("account lookup failed")

type AccountSnapshot struct {
	AccountCountry string
	HomeCurrency   string
	BalanceMinor   int64
	AccountAgeDays int64
}

type RateLimitResult struct {
	Exceeded                 bool
	DailyUtilization         float64
	DailyLimitExceeded       bool
	SenderTxCount24H         int64
	SenderTotalAmountUSD24H  float64
	ReceiverInboundCount24H  int64
	SenderReceiverTxCount24H int64
}
