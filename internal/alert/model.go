package alert

import "time"

const (
	ServiceName = "alert-service"
	DefaultPort = 8086
	InputTopic  = "fraud.final"
)

type TransferStatus string

const (
	TransferStatusProcessing      TransferStatus = "PROCESSING"
	TransferStatusApplied         TransferStatus = "APPLIED"
	TransferStatusFailedRetryable TransferStatus = "FAILED_RETRYABLE"
)

type AlertRecord struct {
	ID             int64     `json:"id"`
	TransactionID  string    `json:"transactionId"`
	Decision       string    `json:"decision"`
	DecisionReason string    `json:"decisionReason"`
	PayloadJSON    string    `json:"payloadJson"`
	DecidedAt      time.Time `json:"decidedAt"`
	CreatedAt      time.Time `json:"createdAt"`
}

type Transfer struct {
	ID            int64          `json:"id"`
	TransactionID string         `json:"transactionId"`
	Status        TransferStatus `json:"status"`
	Attempts      int            `json:"attempts"`
	LastError     *string        `json:"lastError"`
	PayloadJSON   string         `json:"payloadJson"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
}
