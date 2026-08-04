package orchestrator

import (
	"context"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/lghuy05/sentinel-pay/internal/platform/redisutil"
)

func TestAggregationRedisLifecycle(t *testing.T) {
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
	client, err := redisutil.NewClient(ctx, redisutil.Config{Host: host, Port: redisPort})
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	service := &Service{redis: client}
	key := aggregateKey("itest-orchestrator-redis")
	_ = client.Del(ctx, key).Err()

	if err := service.upsertAggregation(ctx, "itest-orchestrator-redis", map[string]string{"ruleBand": "GRAY"}); err != nil {
		t.Fatal(err)
	}
	values, err := client.HGetAll(ctx, key).Result()
	if err != nil {
		t.Fatal(err)
	}
	if values["ruleBand"] != "GRAY" {
		t.Fatalf("ruleBand = %q, want GRAY", values["ruleBand"])
	}
}
