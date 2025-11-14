package postgres

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vsespontanno/eCommerce/cart-service/internal/domain/models"
	"go.uber.org/zap"
)

func TestCartStore_GetCart(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	logger := zap.NewNop().Sugar()
	store := NewCartStore(sqlxDB, logger)

	t.Run("success - cart with items", func(t *testing.T) {
		userID := int64(1)
		rows := sqlmock.NewRows([]string{"id", "user_id", "product_id", "quantity"}).
			AddRow("item1", userID, int64(101), 2).
			AddRow("item2", userID, int64(102), 1)

		mock.ExpectQuery(`SELECT id, user_id, product_id, quantity FROM cart WHERE user_id = \$1`).
			WithArgs(userID).
			WillReturnRows(rows)

		cart, err := store.GetCart(context.Background(), userID)

		require.NoError(t, err)
		require.Len(t, cart.Items, 2)
		assert.Equal(t, int64(101), cart.Items[0].ProductID)
		assert.Equal(t, 2, cart.Items[0].Quantity)
		assert.Equal(t, int64(102), cart.Items[1].ProductID)
		assert.Equal(t, 1, cart.Items[1].Quantity)
	})

	t.Run("success - empty cart", func(t *testing.T) {
		userID := int64(2)
		rows := sqlmock.NewRows([]string{"id", "user_id", "product_id", "quantity"})

		mock.ExpectQuery(`SELECT id, user_id, product_id, quantity FROM cart WHERE user_id = \$1`).
			WithArgs(userID).
			WillReturnRows(rows)

		cart, err := store.GetCart(context.Background(), userID)

		require.NoError(t, err)
		assert.Empty(t, cart.Items)
	})

	t.Run("error - database error", func(t *testing.T) {
		userID := int64(3)

		mock.ExpectQuery(`SELECT id, user_id, product_id, quantity FROM cart WHERE user_id = \$1`).
			WithArgs(userID).
			WillReturnError(errors.New("database error"))

		cart, err := store.GetCart(context.Background(), userID)

		assert.Error(t, err)
		assert.Empty(t, cart.Items)
	})

	t.Run("error - sql no rows", func(t *testing.T) {
		userID := int64(4)

		mock.ExpectQuery(`SELECT id, user_id, product_id, quantity FROM cart WHERE user_id = \$1`).
			WithArgs(userID).
			WillReturnError(sql.ErrNoRows)

		cart, err := store.GetCart(context.Background(), userID)

		assert.Error(t, err)
		assert.Equal(t, models.ErrNoCartFound, err)
		assert.Empty(t, cart.Items)
	})
}

func TestCartStore_UpsertCart(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	logger := zap.NewNop().Sugar()
	store := NewCartStore(sqlxDB, logger)

	query := regexp.QuoteMeta(
		"INSERT INTO cart (user_id,product_id,quantity) VALUES ($1,$2,$3) ON CONFLICT (user_id, product_id) DO UPDATE SET quantity = EXCLUDED.quantity",
	)

	t.Run("success - upsert multiple items", func(t *testing.T) {
		userID := int64(1)
		cartItems := []models.CartItem{
			{ProductID: 101, Quantity: 2},
			{ProductID: 102, Quantity: 1},
		}

		mock.ExpectBegin()
		mock.ExpectExec(query).
			WithArgs(userID, int64(101), 2).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(query).
			WithArgs(userID, int64(102), 1).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := store.UpsertCart(context.Background(), userID, &cartItems)
		assert.NoError(t, err)
	})

	t.Run("error - begin transaction failed", func(t *testing.T) {
		userID := int64(1)
		cartItems := []models.CartItem{{ProductID: 101, Quantity: 2}}

		mock.ExpectBegin().WillReturnError(errors.New("tx begin error"))

		err := store.UpsertCart(context.Background(), userID, &cartItems)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to start transaction")
	})

	t.Run("error - upsert execution failed", func(t *testing.T) {
		userID := int64(1)
		cartItems := []models.CartItem{{ProductID: 101, Quantity: 2}}

		mock.ExpectBegin()
		mock.ExpectExec(query).
			WithArgs(userID, int64(101), 2).
			WillReturnError(errors.New("exec error"))
		mock.ExpectRollback()

		err := store.UpsertCart(context.Background(), userID, &cartItems)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to exec upsert")
	})

	t.Run("error - commit failed", func(t *testing.T) {
		userID := int64(1)
		cartItems := []models.CartItem{{ProductID: 101, Quantity: 2}}

		mock.ExpectBegin()
		mock.ExpectExec(query).
			WithArgs(userID, int64(101), 2).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit().WillReturnError(errors.New("commit error"))

		err := store.UpsertCart(context.Background(), userID, &cartItems)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to commit upsert")
	})
}
func TestCartStore_CleanCart(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	logger := zap.NewNop().Sugar()
	store := NewCartStore(sqlxDB, logger)

	t.Run("success - clean cart with multiple products", func(t *testing.T) {
		order := &models.OrderEvent{
			UserID: 1,
			Products: []models.ProductForOrder{
				{ID: 101, Quantity: 2},
				{ID: 102, Quantity: 1},
			},
		}

		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM cart WHERE user_id = \$1 AND product_id = \$2`).
			WithArgs(int64(1), int64(101)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(`DELETE FROM cart WHERE user_id = \$1 AND product_id = \$2`).
			WithArgs(int64(1), int64(102)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := store.CleanCart(context.Background(), order)

		assert.NoError(t, err)
	})

	t.Run("error - begin transaction failed", func(t *testing.T) {
		order := &models.OrderEvent{
			UserID:   1,
			Products: []models.ProductForOrder{{ID: 101, Quantity: 2}},
		}

		mock.ExpectBegin().WillReturnError(errors.New("tx begin error"))

		err := store.CleanCart(context.Background(), order)

		assert.Error(t, err)
	})

	t.Run("error - delete execution failed", func(t *testing.T) {
		order := &models.OrderEvent{
			UserID:   1,
			Products: []models.ProductForOrder{{ID: 101, Quantity: 2}},
		}

		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM cart WHERE user_id = \$1 AND product_id = \$2`).
			WithArgs(int64(1), int64(101)).
			WillReturnError(errors.New("delete error"))
		mock.ExpectRollback()

		err := store.CleanCart(context.Background(), order)

		assert.Error(t, err)
	})

	t.Run("error - commit failed", func(t *testing.T) {
		order := &models.OrderEvent{
			UserID:   1,
			Products: []models.ProductForOrder{{ID: 101, Quantity: 2}},
		}

		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM cart WHERE user_id = \$1 AND product_id = \$2`).
			WithArgs(int64(1), int64(101)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit().WillReturnError(errors.New("commit error"))

		err := store.CleanCart(context.Background(), order)

		assert.Error(t, err)
	})
}

func TestNewCartStore(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	logger := zap.NewNop().Sugar()

	store := NewCartStore(sqlxDB, logger)

	assert.NotNil(t, store)
	assert.Equal(t, sqlxDB, store.db)
	assert.Equal(t, logger, store.logger)
	assert.NotNil(t, store.builder)

	// Ensure no unexpected database calls
	assert.NoError(t, mock.ExpectationsWereMet())
}
