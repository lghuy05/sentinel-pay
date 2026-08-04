package feature

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/lghuy05/sentinel-pay/internal/contracts"
	"github.com/redis/go-redis/v9"
)

type RedisFeatures struct {
	client *redis.Client
	now    func() time.Time
}

func NewRedisFeatures(client *redis.Client) *RedisFeatures {
	return &RedisFeatures{client: client, now: time.Now}
}

func (r *RedisFeatures) IncrInt(ctx context.Context, key string, delta int64, ttl time.Duration) int64 {
	value, err := r.client.IncrBy(ctx, key, delta).Result()
	if err != nil {
		return 0
	}
	r.ensureTTL(ctx, key, ttl)
	return value
}

func (r *RedisFeatures) IncrFloat(ctx context.Context, key string, delta float64, ttl time.Duration) float64 {
	value, err := r.client.IncrByFloat(ctx, key, delta).Result()
	if err != nil {
		return 0
	}
	r.ensureTTL(ctx, key, ttl)
	return value
}

func (r *RedisFeatures) UniqueMerchantCount(ctx context.Context, senderID int64, merchantID *int64) int64 {
	if merchantID == nil {
		return 0
	}
	key := fmt.Sprintf("feature:unique_merchants_24h:%d", senderID)
	_ = r.client.SAdd(ctx, key, fmt.Sprintf("%d", *merchantID)).Err()
	_ = r.client.Expire(ctx, key, 24*time.Hour).Err()
	count, _ := r.client.SCard(ctx, key).Result()
	return count
}

func (r *RedisFeatures) DeviceState(ctx context.Context, deviceID string) (*time.Time, bool) {
	key := "device_seen:" + deviceID
	raw, err := r.client.Get(ctx, key).Result()
	var previous *time.Time
	if err == nil {
		var millis int64
		if _, scanErr := fmt.Sscanf(raw, "%d", &millis); scanErr == nil {
			t := time.UnixMilli(millis).UTC()
			previous = &t
		}
	}
	_ = r.client.Set(ctx, key, fmt.Sprintf("%d", r.now().UnixMilli()), 7*24*time.Hour).Err()
	return previous, previous == nil
}

func (r *RedisFeatures) LastTxState(ctx context.Context, senderID int64) (*time.Time, *int64) {
	key := fmt.Sprintf("last_tx_time:%d", senderID)
	raw, err := r.client.Get(ctx, key).Result()
	var last *time.Time
	var seconds *int64
	if err == nil {
		var millis int64
		if _, scanErr := fmt.Sscanf(raw, "%d", &millis); scanErr == nil {
			t := time.UnixMilli(millis).UTC()
			value := int64(r.now().Sub(t).Seconds())
			last = &t
			seconds = &value
		}
	}
	_ = r.client.Set(ctx, key, fmt.Sprintf("%d", r.now().UnixMilli()), 7*24*time.Hour).Err()
	return last, seconds
}

func (r *RedisFeatures) FirstTimeContact(ctx context.Context, senderID int64, receiverID *int64, merchantID *int64) bool {
	contact := ""
	if receiverID != nil {
		contact = fmt.Sprintf("user:%d", *receiverID)
	} else if merchantID != nil {
		contact = fmt.Sprintf("merchant:%d", *merchantID)
	}
	if contact == "" {
		return false
	}
	key := fmt.Sprintf("feature:first_contact:%d", senderID)
	added, err := r.client.SAdd(ctx, key, contact).Result()
	if err != nil {
		return false
	}
	_ = r.client.Expire(ctx, key, 30*24*time.Hour).Err()
	return added == 1
}

func (r *RedisFeatures) RateLimit(ctx context.Context, senderID int64, receiverID *int64, txType contracts.TransactionType, eventTime time.Time, amountUSD float64, crossBorder bool) RateLimitResult {
	hourBucket := eventTime.UTC().Format("2006010215")
	dayBucket := eventTime.UTC().Format("20060102")
	hourCount := r.IncrInt(ctx, fmt.Sprintf("rate:tx:hour:%d:%s", senderID, hourBucket), 1, 2*time.Hour)
	dayCount := r.IncrInt(ctx, fmt.Sprintf("rate:tx:day:%d:%s", senderID, dayBucket), 1, 48*time.Hour)
	segment := "domestic"
	dailyLimit := int64(50_000_000)
	if crossBorder {
		segment = "cross"
		dailyLimit = 10_000_000
	}
	totalAmountVND := r.IncrInt(ctx, fmt.Sprintf("rate:amt:day:%s:%d:%s", segment, senderID, dayBucket), int64(math.Round(amountUSD*25000)), 48*time.Hour)
	totalAmountUSD := r.IncrFloat(ctx, fmt.Sprintf("rate:amt:usd:day:%d:%s", senderID, dayBucket), amountUSD, 48*time.Hour)
	pairCount := int64(0)
	inboundCount := int64(0)
	if receiverID != nil {
		pairCount = r.IncrInt(ctx, fmt.Sprintf("rate:pair:day:%d:%d:%s", senderID, *receiverID, dayBucket), 1, 48*time.Hour)
		inboundCount = r.IncrInt(ctx, fmt.Sprintf("rate:inbound:day:%d:%s", *receiverID, dayBucket), 1, 48*time.Hour)
	}
	dailyTxLimit := int64(20)
	if txType == contracts.TransactionTypeMerchantPayment {
		dailyTxLimit = 60
	}
	exceeded := hourCount > 10 || dayCount > dailyTxLimit || totalAmountVND > dailyLimit || pairCount > 3
	utilization := 0.0
	if dailyLimit > 0 {
		utilization = math.Min(1.5, float64(totalAmountVND)/float64(dailyLimit))
	}
	return RateLimitResult{Exceeded: exceeded, DailyUtilization: utilization, DailyLimitExceeded: totalAmountVND > dailyLimit, SenderTxCount24H: dayCount, SenderTotalAmountUSD24H: totalAmountUSD, ReceiverInboundCount24H: inboundCount, SenderReceiverTxCount24H: pairCount}
}

func (r *RedisFeatures) ensureTTL(ctx context.Context, key string, ttl time.Duration) {
	current, err := r.client.TTL(ctx, key).Result()
	if err == nil && current < 0 {
		_ = r.client.Expire(ctx, key, ttl).Err()
	}
}
