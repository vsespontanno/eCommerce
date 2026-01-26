package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type CartClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewCartClient(baseURL string) *CartClient {
	return &CartClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

type CartResponse struct {
	Items []CartItem `json:"items"`
}

type CartItem struct {
	UserID    int64 `json:"user_id"`
	ProductID int64 `json:"product_id"`
	Quantity  int64 `json:"quantity"`
	Price     int64 `json:"price"`
}

type CheckoutResponse struct {
	OrderID string `json:"orderId"`
	Message string `json:"message"`
}

func (c *CartClient) GetCart(ctx context.Context, token string) (*CartResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/cart", nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get cart failed with status: %d", resp.StatusCode)
	}

	var cartResp CartResponse
	if err := json.NewDecoder(resp.Body).Decode(&cartResp); err != nil {
		return nil, err
	}

	return &cartResp, nil
}

func (c *CartClient) IncrementProduct(ctx context.Context, token string, productID int64) error {
	url := fmt.Sprintf("%s/cart/%d/increment", c.baseURL, productID)
	httpReq, err := http.NewRequestWithContext(ctx, "PATCH", url, nil)
	if err != nil {
		return err
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("increment product failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (c *CartClient) DecrementProduct(ctx context.Context, token string, productID int64) error {
	url := fmt.Sprintf("%s/cart/%d/decrement", c.baseURL, productID)
	httpReq, err := http.NewRequestWithContext(ctx, "PATCH", url, nil)
	if err != nil {
		return err
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("decrement product failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (c *CartClient) RemoveProduct(ctx context.Context, token string, productID int64) error {
	url := fmt.Sprintf("%s/cart/%d", c.baseURL, productID)
	httpReq, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return err
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("remove product failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (c *CartClient) ClearCart(ctx context.Context, token string) error {
	httpReq, err := http.NewRequestWithContext(ctx, "DELETE", c.baseURL+"/cart", nil)
	if err != nil {
		return err
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("clear cart failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (c *CartClient) Checkout(ctx context.Context, token string) (*CheckoutResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/cart/order/checkout", nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("checkout failed with status: %d", resp.StatusCode)
	}

	var checkoutResp CheckoutResponse
	if err := json.NewDecoder(resp.Body).Decode(&checkoutResp); err != nil {
		return nil, err
	}

	return &checkoutResp, nil
}
