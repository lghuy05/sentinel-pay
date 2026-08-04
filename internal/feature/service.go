package feature

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/lghuy05/sentinel-pay/internal/contracts"
)

type Service struct {
	accounts *AccountClient
	redis    *RedisFeatures
	now      func() time.Time
}

func NewService(accounts *AccountClient, redisFeatures *RedisFeatures) *Service {
	return &Service{accounts: accounts, redis: redisFeatures, now: time.Now}
}

func (s *Service) Extract(ctx context.Context, event contracts.TransactionReceivedEvent) contracts.TransactionEnrichedEvent {
	senderID := event.SenderUserID
	senderPtr := &senderID
	sender := s.accounts.Fetch(ctx, senderPtr)
	receiver := s.accounts.Fetch(ctx, event.ReceiverUserID)
	senderCountry := normalizeCountry(sender.AccountCountry)
	receiverCountry := normalizeCountry(receiver.AccountCountry)
	crossBorder := isCrossBorder(senderCountry, receiverCountry)
	amountUSD := toUSDEquivalent(event.Amount, event.Currency)
	riskTier := amountRiskTier(amountUSD)
	deviceID := event.DeviceID
	if deviceID == "" {
		deviceID = "unknown"
	}

	txCount1m := s.redis.IncrInt(ctx, "velocity:"+itoa(senderID)+":1m", 1, time.Minute)
	txCount1h := s.redis.IncrInt(ctx, "velocity:"+itoa(senderID)+":1h", 1, time.Hour)
	amount1h := s.redis.IncrFloat(ctx, "amount:"+itoa(senderID)+":1h", event.Amount, time.Hour)
	uniqueMerchants := s.redis.UniqueMerchantCount(ctx, senderID, event.MerchantID)
	deviceLastSeen, newDevice := s.redis.DeviceState(ctx, deviceID)
	lastTxTime, timeSinceLastTx := s.redis.LastTxState(ctx, senderID)
	firstContact := s.redis.FirstTimeContact(ctx, senderID, event.ReceiverUserID, event.MerchantID)
	rateLimit := s.redis.RateLimit(ctx, senderID, event.ReceiverUserID, event.Type, event.EventTime, amountUSD, crossBorder)
	riskFlag := buildRiskFlag(rateLimit.Exceeded, event.Currency)

	return contracts.TransactionEnrichedEvent{
		TransactionID:            event.TransactionID,
		Type:                     event.Type,
		SenderUserID:             event.SenderUserID,
		ReceiverUserID:           event.ReceiverUserID,
		MerchantID:               event.MerchantID,
		Amount:                   event.Amount,
		Currency:                 event.Currency,
		DeviceID:                 deviceID,
		EventTime:                event.EventTime,
		ReceivedAt:               event.ReceivedAt,
		SenderAccountCountry:     senderCountry,
		ReceiverAccountCountry:   receiverCountry,
		SenderCountry:            senderCountry,
		ReceiverCountry:          receiverCountry,
		SenderHomeCurrency:       sender.HomeCurrency,
		SenderBalanceMinor:       sender.BalanceMinor,
		RiskFlag:                 riskFlag,
		DailyAmountUtilization:   rateLimit.DailyUtilization,
		DailyLimitExceeded:       rateLimit.DailyLimitExceeded,
		SenderAccountAgeDays:     sender.AccountAgeDays,
		ReceiverAccountAgeDays:   receiver.AccountAgeDays,
		SenderTxCount24H:         rateLimit.SenderTxCount24H,
		SenderTotalAmountUSD24H:  rateLimit.SenderTotalAmountUSD24H,
		ReceiverInboundCount24H:  rateLimit.ReceiverInboundCount24H,
		UniqueMerchantsLast24H:   uniqueMerchants,
		DeviceLastSeenAt:         deviceLastSeen,
		FeatureComputedAt:        s.now().UTC(),
		CrossBorder:              crossBorder,
		Overseas:                 crossBorder,
		AmountRiskTier:           riskTier,
		FirstTimeContact:         firstContact,
		AmountUSDEquivalent:      amountUSD,
		SenderReceiverTxCount24H: rateLimit.SenderReceiverTxCount24H,
		TxCountLast1Min:          txCount1m,
		TxCountLast1Hour:         txCount1h,
		TxAmountLast1Hour:        amount1h,
		AmountLast1Hour:          amount1h,
		LastTxTime:               lastTxTime,
		TimeSinceLastTxSeconds:   timeSinceLastTx,
		NewDevice:                newDevice,
	}
}

func normalizeCountry(country string) string {
	country = strings.TrimSpace(country)
	if country == "" {
		return "UNKNOWN"
	}
	return strings.ToUpper(country)
}

func isCrossBorder(sender, receiver string) bool {
	if sender == "" || receiver == "" || strings.EqualFold(sender, "UNKNOWN") || strings.EqualFold(receiver, "UNKNOWN") {
		return false
	}
	return !strings.EqualFold(sender, receiver)
}

func toUSDEquivalent(amount float64, currency string) float64 {
	switch strings.ToUpper(currency) {
	case "USD":
		return amount
	case "VND":
		return amount / 25000
	case "EUR":
		return amount * 1.1
	case "JPY":
		return amount * 0.007
	default:
		return amount
	}
}

func amountRiskTier(amountUSD float64) contracts.AmountRiskTier {
	if amountUSD < 50 {
		return contracts.AmountRiskTierLow
	}
	if amountUSD < 300 {
		return contracts.AmountRiskTierMedium
	}
	if amountUSD < 1000 {
		return contracts.AmountRiskTierHigh
	}
	return contracts.AmountRiskTierCritical
}

func buildRiskFlag(exceeded bool, currency string) *string {
	parts := make([]string, 0, 2)
	if exceeded {
		parts = append(parts, "RATE_LIMIT_EXCEEDED")
	}
	switch strings.ToUpper(currency) {
	case "USD", "VND", "EUR", "JPY":
	default:
		parts = append(parts, "UNKNOWN_CURRENCY")
	}
	if len(parts) == 0 {
		return nil
	}
	value := strings.Join(parts, ",")
	return &value
}

func itoa(value int64) string {
	return strconv.FormatInt(value, 10)
}
