package contracts

import "time"

type TransactionType string

const (
	TransactionTypeMerchantPayment TransactionType = "MERCHANT_PAYMENT"
	TransactionTypeP2PTransfer     TransactionType = "P2P_TRANSFER"
)

type CreateTransactionRequest struct {
	TransactionID  string          `json:"transactionId"`
	Type           TransactionType `json:"type"`
	SenderUserID   int64           `json:"senderUserId"`
	Amount         float64         `json:"amount"`
	Currency       string          `json:"currency"`
	Timestamp      time.Time       `json:"timestamp"`
	MerchantID     *int64          `json:"merchantId"`
	ReceiverUserID *int64          `json:"receiverUserId"`
	DeviceID       string          `json:"deviceId"`
}

type TransactionReceivedEvent struct {
	TransactionID  string          `json:"transactionId"`
	Type           TransactionType `json:"type"`
	SenderUserID   int64           `json:"senderUserId"`
	ReceiverUserID *int64          `json:"receiverUserId"`
	MerchantID     *int64          `json:"merchantId"`
	Amount         float64         `json:"amount"`
	Currency       string          `json:"currency"`
	DeviceID       string          `json:"deviceId"`
	EventTime      time.Time       `json:"eventTime"`
	ReceivedAt     time.Time       `json:"receivedAt"`
}
