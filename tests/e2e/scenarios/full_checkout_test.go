package scenarios

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/vsespontanno/eCommerce/tests/e2e/clients"
	"github.com/vsespontanno/eCommerce/tests/e2e/config"
)

type FullCheckoutSuite struct {
	suite.Suite
	cfg            *config.E2EConfig
	ssoClient      *clients.SSOClient
	walletClient   *clients.WalletClient
	productsClient *clients.ProductsClient
	cartClient     *clients.CartClient
	orderClient    *clients.OrderClient
}

func (s *FullCheckoutSuite) SetupSuite() {
	s.cfg = config.LoadConfig()
	s.ssoClient = clients.NewSSOClient(s.cfg.SSOBaseURL)
	s.walletClient = clients.NewWalletClient(s.cfg.WalletBaseURL)
	s.productsClient = clients.NewProductsClient(s.cfg.ProductsBaseURL)
	s.cartClient = clients.NewCartClient(s.cfg.CartBaseURL)

	var err error
	s.orderClient, err = clients.NewOrderClient(s.cfg.OrderGRPCAddr)
	s.Require().NoError(err, "failed to create order client")
}

func (s *FullCheckoutSuite) TearDownSuite() {
	if s.orderClient != nil {
		s.orderClient.Close()
	}
}

func (s *FullCheckoutSuite) TestFullCheckoutFlow() {
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.RequestTimeout*5)
	defer cancel()
	t := time.Now().Unix()
	timestamp := strconv.Itoa(int(t))
	// 1. Register and Login
	email := fmt.Sprintf("testuser%d@example.com", t)
	password := "password123" + timestamp
	firstName := "Test"
	lastName := "User"

	err := s.ssoClient.Register(ctx, clients.RegisterRequest{
		Email:     email,
		Password:  password,
		FirstName: firstName,
		LastName:  lastName,
	})
	s.Require().NoError(err, "failed to register user")

	loginResp, err := s.ssoClient.Login(ctx, email, password)
	s.Require().NoError(err, "failed to login")
	s.Require().NotEmpty(loginResp.Token, "token should not be empty")
	token := loginResp.Token

	// 2. Create Wallet and TopUp
	err = s.walletClient.CreateWallet(ctx, token)
	s.Require().NoError(err, "failed to create wallet")

	initialBalance := int64(10000) // 100.00
	err = s.walletClient.TopUp(ctx, token, initialBalance)
	s.Require().NoError(err, "failed to top up wallet")

	balance, err := s.walletClient.GetBalance(ctx, token)
	s.Require().NoError(err, "failed to get balance")
	s.Equal(initialBalance, balance, "balance mismatch after top up")

	// 3. Get Products
	products, err := s.productsClient.GetProducts(ctx)
	s.Require().NoError(err, "failed to get products")
	s.Require().NotEmpty(products, "products list should not be empty")

	targetProduct := products[0]
	s.T().Logf("Selected product: ID=%d, Name=%s, Price=%d", targetProduct.ID, targetProduct.Name, targetProduct.Price)

	// 4. Add to Cart and Modify Quantity
	// Add product (quantity = 1)
	err = s.productsClient.AddProductToCart(ctx, token, targetProduct.ID)
	s.Require().NoError(err, "failed to add product to cart")

	// Sync Cart (Load from Postgres to Redis if needed)
	_, err = s.cartClient.GetCart(ctx, token)
	s.Require().NoError(err, "failed to sync cart")

	// Increment (quantity = 2)
	err = s.cartClient.IncrementProduct(ctx, token, targetProduct.ID)
	s.Require().NoError(err, "failed to increment product")

	// Increment (quantity = 3)
	err = s.cartClient.IncrementProduct(ctx, token, targetProduct.ID)
	s.Require().NoError(err, "failed to increment product")

	// Decrement (quantity = 2)
	err = s.cartClient.DecrementProduct(ctx, token, targetProduct.ID)
	s.Require().NoError(err, "failed to decrement product")

	// Verify Cart State
	cart, err := s.cartClient.GetCart(ctx, token)
	s.Require().NoError(err, "failed to get cart")
	s.Require().Len(cart.Items, 1, "cart should have 1 item type")
	s.Equal(targetProduct.ID, cart.Items[0].ProductID, "product ID mismatch")
	s.Equal(int64(2), cart.Items[0].Quantity, "quantity mismatch")

	expectedTotal := targetProduct.Price * 2

	// 5. Checkout
	checkoutResp, err := s.cartClient.Checkout(ctx, token)
	s.Require().NoError(err, "checkout failed")
	s.Require().NotEmpty(checkoutResp.OrderID, "order ID should not be empty")
	orderID := checkoutResp.OrderID
	s.T().Logf("Checkout successful, OrderID: %s", orderID)

	// 6. Verify Order Status (Polling)
	s.Eventually(func() bool {
		order, err := s.orderClient.GetOrder(ctx, orderID)
		if err != nil {
			return false
		}
		s.T().Logf("Order Status: %s", order.Status)
		return order.Status == "Completed" // Case sensitive check
	}, s.cfg.PollTimeout, s.cfg.PollInterval, "order status did not become completed")

	// 7. Verify Wallet Balance
	finalBalance, err := s.walletClient.GetBalance(ctx, token)
	s.Require().NoError(err, "failed to get final balance")
	s.Equal(initialBalance-expectedTotal, finalBalance, "final balance mismatch")

	// 8. Verify Cart is Empty
	cartAfter, err := s.cartClient.GetCart(ctx, token)
	if err == nil {
		s.Empty(cartAfter.Items, "cart should be empty after checkout")
	} else {
		s.T().Logf("GetCart returned error (expected if 404 for empty): %v", err)
	}
}

func TestFullCheckoutSuite(t *testing.T) {
	suite.Run(t, new(FullCheckoutSuite))
}
