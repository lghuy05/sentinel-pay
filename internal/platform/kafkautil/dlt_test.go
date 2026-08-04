package kafkautil

import (
	"errors"
	"testing"

	"github.com/segmentio/kafka-go"
)

func TestDLTMessagePreservesPayloadAndAddsErrorHeader(t *testing.T) {
	original := kafka.Message{
		Topic: "transactions.raw",
		Key:   []byte("tx-1"),
		Value: []byte("{bad"),
		Headers: []kafka.Header{
			{Key: "source", Value: []byte("test")},
		},
	}

	got := DLTMessage(original, errors.New("decode failed"))

	if got.Topic != "transactions.raw.DLT" {
		t.Fatalf("Topic = %q, want transactions.raw.DLT", got.Topic)
	}
	if string(got.Key) != "tx-1" || string(got.Value) != "{bad" {
		t.Fatalf("payload not preserved: key=%q value=%q", string(got.Key), string(got.Value))
	}
	if len(got.Headers) != 2 {
		t.Fatalf("headers = %d, want 2", len(got.Headers))
	}
	if got.Headers[1].Key != DLTHeaderError || string(got.Headers[1].Value) != "decode failed" {
		t.Fatalf("error header = %#v", got.Headers[1])
	}
	if len(original.Headers) != 1 {
		t.Fatalf("original headers mutated: %#v", original.Headers)
	}
}
