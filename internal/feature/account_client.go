package feature

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type AccountClient struct {
	baseURL string
	client  *http.Client
}

func NewAccountClient(baseURL string) *AccountClient {
	return &AccountClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: 3 * time.Second},
	}
}

func (c *AccountClient) Fetch(ctx context.Context, userID *int64) AccountSnapshot {
	snapshot := AccountSnapshot{AccountCountry: "UNKNOWN"}
	if userID == nil {
		return snapshot
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/api/v1/accounts/%d", c.baseURL, *userID), nil)
	if err != nil {
		return snapshot
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return snapshot
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return snapshot
	}
	var payload struct {
		AccountCountry string    `json:"accountCountry"`
		HomeCurrency   string    `json:"homeCurrency"`
		CreatedAt      time.Time `json:"createdAt"`
		BalanceMinor   int64     `json:"balanceMinor"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return snapshot
	}
	snapshot.AccountCountry = payload.AccountCountry
	snapshot.HomeCurrency = payload.HomeCurrency
	snapshot.BalanceMinor = payload.BalanceMinor
	if !payload.CreatedAt.IsZero() {
		snapshot.AccountAgeDays = max(0, int64(time.Since(payload.CreatedAt).Hours()/24))
	}
	return snapshot
}
