package redisutil

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestNewClientFailsWhenRedisUnavailable(t *testing.T) {
	t.Parallel()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := listener.Addr().(*net.TCPAddr)
	_ = listener.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	_, err = NewClient(ctx, Config{
		Host:        "127.0.0.1",
		Port:        addr.Port,
		DialTimeout: 200 * time.Millisecond,
		ReadTimeout: 200 * time.Millisecond,
	})
	if err == nil {
		t.Fatal("expected redis connection failure")
	}
}
