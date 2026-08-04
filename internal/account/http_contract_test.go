package account

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAccountHTTPResponseMatchesFixture(t *testing.T) {
	t.Parallel()

	store := &fixtureStore{
		account: Account{
			UserID:         101,
			AccountCountry: "US",
			HomeCurrency:   "USD",
			CreatedAt:      time.Date(2026, 8, 3, 7, 0, 0, 0, time.UTC),
			KycLevel:       "FULL",
			Status:         "ACTIVE",
			BalanceMinor:   125000,
		},
	}

	handler := NewHandler(NewService(store))
	mux := http.NewServeMux()
	handler.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/101", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var got any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}

	want := readFixture(t, "account-response.json")
	if !jsonEqual(got, want) {
		t.Fatalf("response mismatch\nwant: %s\ngot: %s", marshalFixture(t, want), marshalFixture(t, got))
	}
}

type fixtureStore struct {
	account Account
}

func (s *fixtureStore) Create(_ context.Context, account Account) (Account, error) {
	return account, nil
}
func (s *fixtureStore) Get(_ context.Context, _ int64) (Account, error) { return s.account, nil }
func (s *fixtureStore) List(_ context.Context, _, _ int) ([]Account, error) {
	return []Account{s.account}, nil
}
func (s *fixtureStore) Update(_ context.Context, _ int64, _ AccountPatch) (Account, error) {
	return s.account, nil
}
func (s *fixtureStore) TopUp(_ context.Context, _ int64, _ int64) (Account, error) {
	return s.account, nil
}
func (s *fixtureStore) Debit(_ context.Context, _ int64, _ int64) (Account, error) {
	return s.account, nil
}
func (s *fixtureStore) Delete(_ context.Context, _ int64) error { return nil }

func readFixture(t *testing.T, name string) any {
	t.Helper()
	payload, err := os.ReadFile(filepath.Join("..", "..", "testdata", "contracts", name))
	if err != nil {
		t.Fatal(err)
	}
	var value any
	if err := json.Unmarshal(payload, &value); err != nil {
		t.Fatal(err)
	}
	return value
}

func jsonEqual(left, right any) bool {
	return marshalFixture(nil, left) == marshalFixture(nil, right)
}

func marshalFixture(t *testing.T, value any) string {
	payload, err := json.Marshal(value)
	if err != nil {
		if t != nil {
			t.Fatal(err)
		}
		return ""
	}
	return string(payload)
}

var _ Store = (*fixtureStore)(nil)
var _ = sql.ErrNoRows
