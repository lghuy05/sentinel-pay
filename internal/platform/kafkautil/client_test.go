package kafkautil

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestParseBrokers(t *testing.T) {
	t.Parallel()

	got := ParseBrokers(" kafka:9092, localhost:19092 , ", "ignored:9092")
	want := []string{"kafka:9092", "localhost:19092"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseBrokers = %v, want %v", got, want)
	}
}

func TestParseBrokersFallback(t *testing.T) {
	t.Parallel()

	got := ParseBrokers("", "localhost:19092")
	want := []string{"localhost:19092"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseBrokers = %v, want %v", got, want)
	}
}

func TestWaitAfterFetchErrorHonorsContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if WaitAfterFetchError(ctx) {
		t.Fatal("expected false when context is canceled")
	}

	ctx, cancel = context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()
	if !WaitAfterFetchError(ctx) {
		t.Fatal("expected true when timer elapses")
	}
}
