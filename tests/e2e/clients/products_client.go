package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type ProductsClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewProductsClient(baseURL string) *ProductsClient {
	return &ProductsClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

type Product struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Price        int64  `json:"price"`
	Description  string `json:"description"`
	Category     string `json:"category"`
	Brand        string `json:"brand"`
	Rating       int    `json:"rating"`
	NumReviews   int    `json:"num_reviews"`
	CreatedAt    string `json:"created_at"`
	CountInStock int    `json:"count_in_stock"`
}

func (c *ProductsClient) GetProducts(ctx context.Context) ([]Product, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/products", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get products failed with status: %d", resp.StatusCode)
	}

	var products []Product
	if err := json.NewDecoder(resp.Body).Decode(&products); err != nil {
		return nil, err
	}

	return products, nil
}

func (c *ProductsClient) GetProduct(ctx context.Context, id int64) (*Product, error) {
	url := fmt.Sprintf("%s/products/%d", c.baseURL, id)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get product failed with status: %d", resp.StatusCode)
	}

	var product Product
	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		return nil, err
	}

	return &product, nil
}

func (c *ProductsClient) AddProductToCart(ctx context.Context, token string, productID int64) error {
	url := fmt.Sprintf("%s/products/%d/add-to-cart", c.baseURL, productID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("add product to cart failed with status: %d", resp.StatusCode)
	}

	return nil
}
