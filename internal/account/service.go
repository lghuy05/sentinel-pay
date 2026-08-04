package account

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"

	"github.com/lghuy05/sentinel-pay/internal/contracts"
)

const vndPerUSD int64 = 25_000

type Store interface {
	Create(ctx context.Context, account Account) (Account, error)
	Get(ctx context.Context, userID int64) (Account, error)
	List(ctx context.Context, limit, offset int) ([]Account, error)
	Update(ctx context.Context, userID int64, patch AccountPatch) (Account, error)
	TopUp(ctx context.Context, userID int64, amountMinor int64) (Account, error)
	Debit(ctx context.Context, userID int64, amountMinor int64) (Account, error)
	Delete(ctx context.Context, userID int64) error
}

type AccountPatch struct {
	AccountCountry *string
	CreatedAt      *time.Time
	KycLevel       *string
	Status         *string
}

type Service struct {
	store Store
	now   func() time.Time
}

func NewService(store Store) *Service {
	return &Service{
		store: store,
		now:   time.Now,
	}
}

func (s *Service) Create(ctx context.Context, request contracts.CreateAccountRequest) (contracts.AccountResponse, error) {
	if err := validateCreate(request); err != nil {
		return contracts.AccountResponse{}, err
	}

	userID := int64(0)
	if request.UserID != nil {
		userID = *request.UserID
	} else {
		generated, err := generateUserID()
		if err != nil {
			return contracts.AccountResponse{}, fmt.Errorf("generate user id: %w", err)
		}
		userID = generated
	}

	createdAt := s.now()
	if request.CreatedAt != nil {
		createdAt = *request.CreatedAt
	}

	kycLevel := string(request.KycLevel)
	if kycLevel == "" {
		kycLevel = string(contracts.KycLevelBasic)
	}

	status := string(request.Status)
	if status == "" {
		status = string(contracts.AccountStatusActive)
	}

	initialBalance := int64(0)
	if request.InitialBalance != nil {
		if *request.InitialBalance < 0 {
			return contracts.AccountResponse{}, fmt.Errorf("%w: initialBalance must be non-negative", ErrInvalidInput)
		}
		initialBalance = *request.InitialBalance
	}

	created, err := s.store.Create(ctx, Account{
		UserID:         userID,
		AccountCountry: request.AccountCountry,
		HomeCurrency:   request.HomeCurrency,
		CreatedAt:      createdAt,
		KycLevel:       kycLevel,
		Status:         status,
		BalanceMinor:   initialBalance,
	})
	if err != nil {
		return contracts.AccountResponse{}, err
	}

	return toResponse(created), nil
}

func (s *Service) Get(ctx context.Context, userID int64) (contracts.AccountResponse, error) {
	account, err := s.store.Get(ctx, userID)
	if err != nil {
		return contracts.AccountResponse{}, err
	}
	return toResponse(account), nil
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]contracts.AccountResponse, error) {
	size := max(1, min(limit, 200))
	page := max(0, offset/size)
	accounts, err := s.store.List(ctx, size, page*size)
	if err != nil {
		return nil, err
	}

	responses := make([]contracts.AccountResponse, 0, len(accounts))
	for _, account := range accounts {
		responses = append(responses, toResponse(account))
	}
	return responses, nil
}

func (s *Service) Update(ctx context.Context, userID int64, request contracts.UpdateAccountRequest) (contracts.AccountResponse, error) {
	patch := AccountPatch{
		AccountCountry: request.AccountCountry,
		CreatedAt:      request.CreatedAt,
	}
	if request.KycLevel != "" {
		value := string(request.KycLevel)
		if !validKycLevel(value) {
			return contracts.AccountResponse{}, fmt.Errorf("%w: invalid kycLevel", ErrInvalidInput)
		}
		patch.KycLevel = &value
	}
	if request.Status != "" {
		value := string(request.Status)
		if !validStatus(value) {
			return contracts.AccountResponse{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
		}
		patch.Status = &value
	}

	account, err := s.store.Update(ctx, userID, patch)
	if err != nil {
		return contracts.AccountResponse{}, err
	}
	return toResponse(account), nil
}

func (s *Service) TopUp(ctx context.Context, userID int64, request contracts.BalanceRequest) (contracts.AccountResponse, error) {
	if request.Amount == nil || *request.Amount <= 0 {
		return contracts.AccountResponse{}, fmt.Errorf("%w: amount must be positive", ErrInvalidInput)
	}
	if strings.TrimSpace(request.Currency) == "" {
		return contracts.AccountResponse{}, fmt.Errorf("%w: currency is required", ErrInvalidInput)
	}

	account, err := s.store.Get(ctx, userID)
	if err != nil {
		return contracts.AccountResponse{}, err
	}
	delta, err := convertToMinor(*request.Amount, request.Currency, account.HomeCurrency)
	if err != nil {
		return contracts.AccountResponse{}, err
	}
	updated, err := s.store.TopUp(ctx, userID, delta)
	if err != nil {
		return contracts.AccountResponse{}, err
	}
	return toResponse(updated), nil
}

func (s *Service) Debit(ctx context.Context, userID int64, request contracts.BalanceRequest) (contracts.AccountResponse, error) {
	if request.Amount == nil || *request.Amount <= 0 {
		return contracts.AccountResponse{}, fmt.Errorf("%w: amount must be positive", ErrInvalidInput)
	}
	if strings.TrimSpace(request.Currency) == "" {
		return contracts.AccountResponse{}, fmt.Errorf("%w: currency is required", ErrInvalidInput)
	}

	account, err := s.store.Get(ctx, userID)
	if err != nil {
		return contracts.AccountResponse{}, err
	}
	delta, err := convertToMinor(*request.Amount, request.Currency, account.HomeCurrency)
	if err != nil {
		return contracts.AccountResponse{}, err
	}
	updated, err := s.store.Debit(ctx, userID, delta)
	if err != nil {
		return contracts.AccountResponse{}, err
	}
	return toResponse(updated), nil
}

func (s *Service) Delete(ctx context.Context, userID int64) error {
	return s.store.Delete(ctx, userID)
}

func validateCreate(request contracts.CreateAccountRequest) error {
	if len(strings.TrimSpace(request.AccountCountry)) != 2 {
		return fmt.Errorf("%w: accountCountry must be exactly 2 characters", ErrInvalidInput)
	}
	if len(strings.TrimSpace(request.HomeCurrency)) != 3 {
		return fmt.Errorf("%w: homeCurrency must be exactly 3 characters", ErrInvalidInput)
	}
	if request.KycLevel != "" && !validKycLevel(string(request.KycLevel)) {
		return fmt.Errorf("%w: invalid kycLevel", ErrInvalidInput)
	}
	if request.Status != "" && !validStatus(string(request.Status)) {
		return fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}
	return nil
}

func validKycLevel(value string) bool {
	return value == string(contracts.KycLevelBasic) || value == string(contracts.KycLevelFull)
}

func validStatus(value string) bool {
	return value == string(contracts.AccountStatusActive) || value == string(contracts.AccountStatusLocked)
}

func convertToMinor(amount int64, currency string, targetCurrency string) (int64, error) {
	from := strings.ToUpper(currency)
	to := strings.ToUpper(targetCurrency)
	if from == "" || to == "" || from == to {
		return amount, nil
	}
	if from == "USD" && to == "VND" {
		if amount > math.MaxInt64/vndPerUSD || amount < math.MinInt64/vndPerUSD {
			return 0, fmt.Errorf("%w: amount out of range", ErrBalanceOverflow)
		}
		return amount * vndPerUSD, nil
	}
	if from == "VND" && to == "USD" {
		return halfUpDivide(amount, vndPerUSD), nil
	}
	return amount, nil
}

func halfUpDivide(value, divisor int64) int64 {
	if value >= 0 {
		return (value + divisor/2) / divisor
	}
	return (value - divisor/2) / divisor
}

func generateUserID() (int64, error) {
	value, err := rand.Int(rand.Reader, big.NewInt(899_999))
	if err != nil {
		var buffer [8]byte
		if _, readErr := rand.Read(buffer[:]); readErr != nil {
			return 0, err
		}
		return int64(binary.BigEndian.Uint64(buffer[:])%899_999) + 100_000, nil
	}
	return value.Int64() + 100_000, nil
}

func toResponse(account Account) contracts.AccountResponse {
	return contracts.AccountResponse{
		UserID:         account.UserID,
		AccountCountry: account.AccountCountry,
		HomeCurrency:   account.HomeCurrency,
		CreatedAt:      account.CreatedAt,
		KycLevel:       contracts.KycLevel(account.KycLevel),
		Status:         contracts.AccountStatus(account.Status),
		BalanceMinor:   account.BalanceMinor,
	}
}
