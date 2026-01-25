package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type WalletClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewWalletClient(baseURL string) *WalletClient {
	return &WalletClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

type TopUpRequest struct {
	Amount int64 `json:"amount"`
}

type BalanceResponse struct {
	Balance int64 `json:"balance"`
}

func (c *WalletClient) CreateWallet(ctx context.Context, token string) error {
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/wallet", nil)
	if err != nil {
		return err
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("create wallet failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (c *WalletClient) TopUp(ctx context.Context, token string, amount int64) error {
	req := TopUpRequest{Amount: amount}
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/wallet/topup", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("top up failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (c *WalletClient) GetBalance(ctx context.Context, token string) (int64, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/wallet/balance", nil)
	if err != nil {
		return 0, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("get balance failed with status: %d", resp.StatusCode)
	}

	var balanceResp BalanceResponse
	if err := json.NewDecoder(resp.Body).Decode(&balanceResp); err != nil {
		return 0, err
	}

	return balanceResp.Balance, nil
}
