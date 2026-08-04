package feature

import (
	"context"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/lghuy05/sentinel-pay/internal/contracts"
	"github.com/lghuy05/sentinel-pay/internal/platform/redisutil"
)

func TestRedisFeaturesIntegration(t *testing.T) {
	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("TEST_REDIS_ADDR not set")
	}
	host, port, ok := strings.Cut(addr, ":")
	if !ok {
		t.Fatalf("invalid TEST_REDIS_ADDR: %s", addr)
	}
	redisPort, err := strconv.Atoi(port)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	client, err := redisutil.NewClient(ctx, redisutil.Config{
		Host: host,
		Port: redisPort,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	features := NewRedisFeatures(client)
	senderID := int64(7001)
	receiverID := int64(7002)
	merchantID := int64(7003)
	_ = client.Del(ctx,
		"device_seen:itest-device",
		"last_tx_time:7001",
		"feature:unique_merchants_24h:7001",
		"feature:first_contact:7001",
		"rate:tx:hour:7001:2026080404",
		"rate:tx:day:7001:20260804",
		"rate:amt:day:domestic:7001:20260804",
		"rate:amt:usd:day:7001:20260804",
		"rate:pair:day:7001:7002:20260804",
		"rate:inbound:day:7002:20260804",
	).Err()

	if count := features.UniqueMerchantCount(ctx, senderID, &merchantID); count != 1 {
		t.Fatalf("unique merchants = %d, want 1", count)
	}
	if first := features.FirstTimeContact(ctx, senderID, &receiverID, nil); !first {
		t.Fatal("expected first contact to be true")
	}
	if first := features.FirstTimeContact(ctx, senderID, &receiverID, nil); first {
		t.Fatal("expected repeated contact to be false")
	}

	previous, isNew := features.DeviceState(ctx, "itest-device")
	if previous != nil || !isNew {
		t.Fatalf("first device state previous=%v isNew=%v", previous, isNew)
	}
	previous, isNew = features.DeviceState(ctx, "itest-device")
	if previous == nil || isNew {
		t.Fatalf("second device state previous=%v isNew=%v", previous, isNew)
	}

	result := features.RateLimit(ctx, senderID, &receiverID, contracts.TransactionTypeP2PTransfer, time.Date(2026, 8, 4, 4, 0, 0, 0, time.UTC), 25, false)
	if result.SenderTxCount24H != 1 {
		t.Fatalf("sender tx count = %d, want 1", result.SenderTxCount24H)
	}
}
