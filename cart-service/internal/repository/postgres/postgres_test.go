package postgres

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/vsespontanno/eCommerce/cart-service/internal/config"
	"github.com/vsespontanno/eCommerce/cart-service/internal/repository"
)

func Init() {

}

func TestGetCart(t *testing.T) {
	cfg, err := config.MustLoad()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	PG_USER := cfg.PGUser
	PG_PASSWORD := cfg.PGPassword
	PG_NAME := cfg.PGName
	PG_HOST := cfg.PGHost
	PG_PORT := cfg.PGPort
	db, err := repository.ConnectToPostgres(PG_USER, PG_PASSWORD, PG_NAME, PG_HOST, PG_PORT)
	store := NewCartStore(db)
	if err != nil {
		t.Errorf("failed to connect to Postgres: %v", err)
	}
	userID := int64(18)
	cart, err := store.GetCart(context.Background(), userID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(cart.Items) == 0 {
		t.Fatalf("expected cart to have items, got none")
	} else {
		fmt.Println(cart.Items)
	}
}
