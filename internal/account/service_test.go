package account

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lghuy05/sentinel-pay/internal/contracts"
)

func TestCreateDefaultsMatchJavaService(t *testing.T) {
	store := newMemoryStore()
	service := NewService(store)
	service.now = func() time.Time {
		return time.Date(2026, 8, 3, 8, 0, 0, 0, time.UTC)
	}
	userID := int64(101)
	initialBalance := int64(100000)

	got, err := service.Create(context.Background(), contracts.CreateAccountRequest{
		UserID:         &userID,
		AccountCountry: "VN",
		HomeCurrency:   "VND",
		InitialBalance: &initialBalance,
	})
	if err != nil {
		t.Fatal(err)
	}

	if got.UserID != userID || got.KycLevel != contracts.KycLevelBasic || got.Status != contracts.AccountStatusActive {
		t.Fatalf("unexpected defaults: %+v", got)
	}
	if got.BalanceMinor != initialBalance {
		t.Fatalf("balanceMinor = %d, want %d", got.BalanceMinor, initialBalance)
	}
}

func TestTopUpConvertsUSDToVND(t *testing.T) {
	store := newMemoryStore()
	service := NewService(store)
	userID := int64(101)
	initialBalance := int64(100)
	_, err := service.Create(context.Background(), contracts.CreateAccountRequest{
		UserID:         &userID,
		AccountCountry: "VN",
		HomeCurrency:   "VND",
		InitialBalance: &initialBalance,
	})
	if err != nil {
		t.Fatal(err)
	}
	amount := int64(2)

	got, err := service.TopUp(context.Background(), userID, contracts.BalanceRequest{
		Amount:   &amount,
		Currency: "USD",
	})
	if err != nil {
		t.Fatal(err)
	}

	if got.BalanceMinor != 50100 {
		t.Fatalf("balanceMinor = %d, want 50100", got.BalanceMinor)
	}
}

func TestDebitRejectsInsufficientFunds(t *testing.T) {
	store := newMemoryStore()
	service := NewService(store)
	userID := int64(101)
	initialBalance := int64(100)
	_, err := service.Create(context.Background(), contracts.CreateAccountRequest{
		UserID:         &userID,
		AccountCountry: "VN",
		HomeCurrency:   "VND",
		InitialBalance: &initialBalance,
	})
	if err != nil {
		t.Fatal(err)
	}
	amount := int64(101)

	_, err = service.Debit(context.Background(), userID, contracts.BalanceRequest{
		Amount:   &amount,
		Currency: "VND",
	})
	if !errors.Is(err, ErrInsufficientFunds) {
		t.Fatalf("err = %v, want ErrInsufficientFunds", err)
	}
}

type memoryStore struct {
	accounts map[int64]Account
}

func newMemoryStore() *memoryStore {
	return &memoryStore{accounts: map[int64]Account{}}
}

func (s *memoryStore) Create(_ context.Context, account Account) (Account, error) {
	s.accounts[account.UserID] = account
	return account, nil
}

func (s *memoryStore) Get(_ context.Context, userID int64) (Account, error) {
	account, ok := s.accounts[userID]
	if !ok {
		return Account{}, ErrNotFound
	}
	return account, nil
}

func (s *memoryStore) List(_ context.Context, _, _ int) ([]Account, error) {
	accounts := make([]Account, 0, len(s.accounts))
	for _, account := range s.accounts {
		accounts = append(accounts, account)
	}
	return accounts, nil
}

func (s *memoryStore) Update(_ context.Context, userID int64, patch AccountPatch) (Account, error) {
	account, ok := s.accounts[userID]
	if !ok {
		return Account{}, ErrNotFound
	}
	if patch.AccountCountry != nil {
		account.AccountCountry = *patch.AccountCountry
	}
	if patch.CreatedAt != nil {
		account.CreatedAt = *patch.CreatedAt
	}
	if patch.KycLevel != nil {
		account.KycLevel = *patch.KycLevel
	}
	if patch.Status != nil {
		account.Status = *patch.Status
	}
	s.accounts[userID] = account
	return account, nil
}

func (s *memoryStore) TopUp(_ context.Context, userID int64, amountMinor int64) (Account, error) {
	account, ok := s.accounts[userID]
	if !ok {
		return Account{}, ErrNotFound
	}
	account.BalanceMinor += amountMinor
	s.accounts[userID] = account
	return account, nil
}

func (s *memoryStore) Debit(_ context.Context, userID int64, amountMinor int64) (Account, error) {
	account, ok := s.accounts[userID]
	if !ok {
		return Account{}, ErrNotFound
	}
	if amountMinor > account.BalanceMinor {
		return Account{}, ErrInsufficientFunds
	}
	account.BalanceMinor -= amountMinor
	s.accounts[userID] = account
	return account, nil
}

func (s *memoryStore) Delete(_ context.Context, userID int64) error {
	if _, ok := s.accounts[userID]; !ok {
		return ErrNotFound
	}
	delete(s.accounts, userID)
	return nil
}
