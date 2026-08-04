package alert

import (
	"bytes"
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
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *AccountClient) Debit(ctx context.Context, userID int64, amount int64, currency string) error {
	return c.postBalance(ctx, userID, "debit", amount, currency)
}

func (c *AccountClient) TopUp(ctx context.Context, userID int64, amount int64, currency string) error {
	return c.postBalance(ctx, userID, "topup", amount, currency)
}

func (c *AccountClient) postBalance(ctx context.Context, userID int64, action string, amount int64, currency string) error {
	payload, err := json.Marshal(map[string]any{"amount": amount, "currency": currency})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/api/v1/accounts/%d/%s", c.baseURL, userID, action), bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("account %s failed for user %d: status %d", action, userID, resp.StatusCode)
	}
	return nil
}
