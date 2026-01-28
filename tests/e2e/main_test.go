package e2e

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/vsespontanno/eCommerce/tests/e2e/config"
)

func TestMain(m *testing.M) {
	cfg := config.LoadConfig()

	// Wait for services to be ready
	if err := waitForServices(cfg); err != nil {
		fmt.Printf("Failed to wait for services: %v\n", err)
		os.Exit(1)
	}

	// Run tests
	code := m.Run()
	os.Exit(code)
}

func waitForServices(cfg *config.E2EConfig) error {
	services := map[string]string{
		"SSO":      cfg.SSOBaseURL + "/health",
		"Wallet":   cfg.WalletBaseURL + "/health",
		"Products": cfg.ProductsBaseURL + "/health",
		"Cart":     cfg.CartBaseURL + "/health",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	for name, url := range services {
		fmt.Printf("Waiting for %s service at %s...\n", name, url)
		if err := waitForURL(ctx, url); err != nil {
			return fmt.Errorf("service %s is not ready: %w", name, err)
		}
		fmt.Printf("Service %s is ready!\n", name)
	}

	return nil
}

func waitForURL(ctx context.Context, url string) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			resp, err := http.Get(url)
			if err == nil && resp.StatusCode == http.StatusOK {
				resp.Body.Close()
				return nil
			}
			if resp != nil {
				resp.Body.Close()
			}
		}
	}
}
