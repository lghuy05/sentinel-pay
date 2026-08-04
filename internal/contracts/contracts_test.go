package contracts

import (
	"encoding/json"
	"testing"
	"time"
)

func TestCreateTransactionRequestJSONMatchesJavaContract(t *testing.T) {
	receiverID := int64(202)
	payload := CreateTransactionRequest{
		TransactionID:  "tx-1",
		Type:           TransactionTypeP2PTransfer,
		SenderUserID:   101,
		ReceiverUserID: &receiverID,
		Amount:         125,
		Currency:       "USD",
		Timestamp:      time.Date(2026, 8, 3, 7, 0, 0, 0, time.UTC),
		DeviceID:       "device-1",
	}

	got, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(got, &decoded); err != nil {
		t.Fatal(err)
	}

	required := []string{"transactionId", "type", "senderUserId", "receiverUserId", "amount", "currency", "timestamp", "deviceId"}
	for _, key := range required {
		if _, ok := decoded[key]; !ok {
			t.Fatalf("missing JSON field %q in %s", key, string(got))
		}
	}
	if decoded["type"] != string(TransactionTypeP2PTransfer) {
		t.Fatalf("type = %v, want %s", decoded["type"], TransactionTypeP2PTransfer)
	}
}

func TestRequiredKafkaTopicsMatchPipeline(t *testing.T) {
	want := map[string]bool{
		"transactions.raw":      true,
		"transactions.enriched": true,
		"fraud.blacklist":       true,
		"fraud.rules":           true,
		"fraud.ml":              true,
		"fraud.final":           true,
	}

	if len(RequiredKafkaTopics) != len(want) {
		t.Fatalf("topic count = %d, want %d", len(RequiredKafkaTopics), len(want))
	}
	for _, topic := range RequiredKafkaTopics {
		if !want[topic] {
			t.Fatalf("unexpected topic %q", topic)
		}
	}
}
