package account

import (
	"errors"
	"time"
)

const (
	ServiceName = "account-service"
	DefaultPort = 8087
)

var (
	ErrNotFound          = errors.New("account not found")
	ErrInvalidInput      = errors.New("invalid account input")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrBalanceOverflow   = errors.New("balance overflow")
)

type Account struct {
	UserID         int64
	AccountCountry string
	HomeCurrency   string
	CreatedAt      time.Time
	KycLevel       string
	Status         string
	BalanceMinor   int64
	Version        int64
}
