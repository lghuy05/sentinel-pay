package ingest

import (
	"errors"
	"time"
)

const (
	ServiceName = "transaction-ingestor"
	DefaultPort = 8081
	RawTopic    = "transactions.raw"
)

var (
	ErrInvalidInput  = errors.New("invalid transaction input")
	ErrNotFound      = errors.New("transaction not found")
	ErrDuplicate     = errors.New("duplicate transactionId")
	ErrRelayDisabled = errors.New("event relay disabled")
)

type TransactionRecord struct {
	ID             int64     `json:"id"`
	TransactionID  string    `json:"transactionId"`
	Type           string    `json:"type"`
	SenderUserID   int64     `json:"senderUserId"`
	ReceiverUserID *int64    `json:"receiverUserId"`
	MerchantID     *int64    `json:"merchantId"`
	Amount         float64   `json:"amount"`
	Currency       string    `json:"currency"`
	DeviceID       string    `json:"deviceId"`
	EventTime      time.Time `json:"eventTime"`
	ReceivedAt     time.Time `json:"receivedAt"`
}

type OutboxEvent struct {
	ID            int64
	AggregateType string
	AggregateID   string
	EventType     string
	Payload       string
	Status        string
	AttemptCount  int
	CreatedAt     time.Time
	NextRetryAt   *time.Time
	PublishedAt   *time.Time
	LastError     *string
}
