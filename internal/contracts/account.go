package contracts

import "time"

type KycLevel string

const (
	KycLevelBasic KycLevel = "BASIC"
	KycLevelFull  KycLevel = "FULL"
)

type AccountStatus string

const (
	AccountStatusActive AccountStatus = "ACTIVE"
	AccountStatusLocked AccountStatus = "LOCKED"
)

type CreateAccountRequest struct {
	UserID         *int64        `json:"userId,omitempty"`
	AccountCountry string        `json:"accountCountry"`
	HomeCurrency   string        `json:"homeCurrency"`
	CreatedAt      *time.Time    `json:"createdAt,omitempty"`
	KycLevel       KycLevel      `json:"kycLevel,omitempty"`
	Status         AccountStatus `json:"status,omitempty"`
	InitialBalance *int64        `json:"initialBalance,omitempty"`
}

type UpdateAccountRequest struct {
	AccountCountry *string       `json:"accountCountry,omitempty"`
	CreatedAt      *time.Time    `json:"createdAt,omitempty"`
	KycLevel       KycLevel      `json:"kycLevel,omitempty"`
	Status         AccountStatus `json:"status,omitempty"`
}

type BalanceRequest struct {
	Amount   *int64 `json:"amount"`
	Currency string `json:"currency"`
}

type AccountResponse struct {
	UserID         int64         `json:"userId"`
	AccountCountry string        `json:"accountCountry"`
	HomeCurrency   string        `json:"homeCurrency"`
	CreatedAt      time.Time     `json:"createdAt"`
	KycLevel       KycLevel      `json:"kycLevel"`
	Status         AccountStatus `json:"status"`
	BalanceMinor   int64         `json:"balanceMinor"`
}
