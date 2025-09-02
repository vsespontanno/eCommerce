package redis

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRedisAddr адрес Redis для тестов
const TestRedisAddr = "6379"

// setupTestRedis создает тестовый клиент Redis
func setupTestRedis(t *testing.T) *redis.Client {
	host := "localhost:" + TestRedisAddr
	rdb := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: "",
		DB:       0,
	})

	// Проверяем соединение с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := rdb.Ping(ctx).Err()
	require.NoError(t, err, "Failed to connect to Redis")

	return rdb
}

// cleanupTestRedis очищает тестовые данные
func cleanupTestRedis(t *testing.T, rdb *redis.Client, keys ...string) {
	ctx := context.Background()
	for _, key := range keys {
		err := rdb.Del(ctx, key).Err()
		assert.NoError(t, err)
	}
}

func TestOrderStore_AddAllQuantityOfProductToCart(t *testing.T) {
	rdb := setupTestRedis(t)
	defer rdb.Close()

	store := NewOrderStore(rdb)
	ctx := context.Background()
	userID := int64(1)
	productID := int64(100)
	quantity := int64(5)
	key := "cart:" + strconv.FormatInt(userID, 10)

	// Очищаем после теста
	defer cleanupTestRedis(t, rdb, key)

	// Тестируем добавление количества
	err := store.AddAllQuantityOfProductToCart(ctx, userID, productID, quantity)
	require.NoError(t, err)

	// Проверяем, что значение установлено правильно
	result, err := rdb.HGet(ctx, key, strconv.FormatInt(productID, 10)).Result()
	require.NoError(t, err)
	assert.Equal(t, "5", result)
}

func TestOrderStore_AddAllProductsToCart(t *testing.T) {
	rdb := setupTestRedis(t)
	defer rdb.Close()

	store := NewOrderStore(rdb)
	ctx := context.Background()
	userID := int64(2)
	productIDs := []int64{101, 102, 103}
	key := "cart:" + strconv.FormatInt(userID, 10)

	defer cleanupTestRedis(t, rdb, key)

	// Тестируем добавление нескольких продуктов
	err := store.AddAllProductsToCart(ctx, userID, productIDs)
	require.NoError(t, err)

	// Проверяем, что все продукты добавлены с количеством 1
	items, err := rdb.HGetAll(ctx, key).Result()
	require.NoError(t, err)

	assert.Len(t, items, 3)
	for _, pid := range productIDs {
		field := strconv.FormatInt(pid, 10)
		assert.Equal(t, "1", items[field])
	}
}

func TestOrderStore_AddToCart(t *testing.T) {
	rdb := setupTestRedis(t)
	defer rdb.Close()

	store := NewOrderStore(rdb)
	ctx := context.Background()
	userID := int64(3)
	productID := int64(200)
	key := "cart:" + strconv.FormatInt(userID, 10)

	defer cleanupTestRedis(t, rdb, key)

	// Тестируем добавление одного продукта
	err := store.AddToCart(ctx, userID, productID)
	require.NoError(t, err)

	// Проверяем, что количество стало 1
	result, err := rdb.HGet(ctx, key, strconv.FormatInt(productID, 10)).Result()
	require.NoError(t, err)
	assert.Equal(t, "1", result)

	// Добавляем еще один раз
	err = store.AddToCart(ctx, userID, productID)
	require.NoError(t, err)

	// Проверяем, что количество стало 2
	result, err = rdb.HGet(ctx, key, strconv.FormatInt(productID, 10)).Result()
	require.NoError(t, err)
	assert.Equal(t, "2", result)
}

func TestOrderStore_RemoveOneFromCart(t *testing.T) {
	rdb := setupTestRedis(t)
	defer rdb.Close()

	store := NewOrderStore(rdb)
	ctx := context.Background()
	userID := int64(4)
	productID := int64(300)
	key := "cart:" + strconv.FormatInt(userID, 10)

	defer cleanupTestRedis(t, rdb, key)

	// Сначала добавляем продукт с количеством 3
	err := store.AddAllQuantityOfProductToCart(ctx, userID, productID, 3)
	require.NoError(t, err)

	// Удаляем один
	err = store.RemoveOneFromCart(ctx, userID, productID)
	require.NoError(t, err)

	// Проверяем, что количество стало 2
	result, err := rdb.HGet(ctx, key, strconv.FormatInt(productID, 10)).Result()
	require.NoError(t, err)
	assert.Equal(t, "2", result)

	// Удаляем еще два раза
	err = store.RemoveOneFromCart(ctx, userID, productID)
	require.NoError(t, err)
	err = store.RemoveOneFromCart(ctx, userID, productID)
	require.NoError(t, err)

	// Проверяем, что поле удалено (количество стало 0 или меньше)
	exists, err := rdb.HExists(ctx, key, strconv.FormatInt(productID, 10)).Result()
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestOrderStore_RemoveProductFromCart(t *testing.T) {
	rdb := setupTestRedis(t)
	defer rdb.Close()

	store := NewOrderStore(rdb)
	ctx := context.Background()
	userID := int64(5)
	productID := int64(400)
	key := "cart:" + strconv.FormatInt(userID, 10)

	defer cleanupTestRedis(t, rdb, key)

	// Добавляем продукт
	err := store.AddAllQuantityOfProductToCart(ctx, userID, productID, 5)
	require.NoError(t, err)

	// Удаляем полностью
	err = store.RemoveProductFromCart(ctx, userID, productID)
	require.NoError(t, err)

	// Проверяем, что поле удалено
	exists, err := rdb.HExists(ctx, key, strconv.FormatInt(productID, 10)).Result()
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestOrderStore_GetCart(t *testing.T) {
	rdb := setupTestRedis(t)
	defer rdb.Close()

	store := NewOrderStore(rdb)
	ctx := context.Background()
	userID := int64(6)
	key := "cart:" + strconv.FormatInt(userID, 10)

	defer cleanupTestRedis(t, rdb, key)

	// Подготавливаем тестовые данные
	testData := map[int64]int64{
		501: 2,
		502: 1,
		503: 5,
	}

	for pid, qty := range testData {
		err := store.AddAllQuantityOfProductToCart(ctx, userID, pid, qty)
		require.NoError(t, err)
	}

	// Получаем корзину
	cart, err := store.GetCart(ctx, userID)
	require.NoError(t, err)

	// Проверяем, что все данные совпадают
	assert.Len(t, cart, 3)
	for pid, expectedQty := range testData {
		assert.Equal(t, expectedQty, cart[pid])
	}
}

func TestOrderStore_GetCart_Empty(t *testing.T) {
	rdb := setupTestRedis(t)
	defer rdb.Close()

	store := NewOrderStore(rdb)
	ctx := context.Background()
	userID := int64(999) // Несуществующий пользователь

	// Получаем пустую корзину
	cart, err := store.GetCart(ctx, userID)
	require.NoError(t, err)
	assert.Empty(t, cart)
}

func TestOrderStore_Integration(t *testing.T) {
	rdb := setupTestRedis(t)
	defer rdb.Close()

	store := NewOrderStore(rdb)
	ctx := context.Background()
	userID := int64(7)
	productID1 := int64(601)
	productID2 := int64(602)
	key := "cart:" + strconv.FormatInt(userID, 10)

	defer cleanupTestRedis(t, rdb, key)

	// Интеграционный тест: последовательность операций
	err := store.AddToCart(ctx, userID, productID1)
	require.NoError(t, err)

	err = store.AddToCart(ctx, userID, productID2)
	require.NoError(t, err)

	err = store.AddToCart(ctx, userID, productID1) // Увеличиваем количество первого продукта
	require.NoError(t, err)

	// Проверяем промежуточное состояние
	cart, err := store.GetCart(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, int64(2), cart[productID1])
	assert.Equal(t, int64(1), cart[productID2])

	// Удаляем один экземпляр первого продукта
	err = store.RemoveOneFromCart(ctx, userID, productID1)
	require.NoError(t, err)

	// Полностью удаляем второй продукт
	err = store.RemoveProductFromCart(ctx, userID, productID2)
	require.NoError(t, err)

	// Проверяем финальное состояние
	cart, err = store.GetCart(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), cart[productID1])
	assert.NotContains(t, cart, productID2)
}

func TestOrderStore_ConcurrentAccess(t *testing.T) {
	rdb := setupTestRedis(t)
	defer rdb.Close()

	store := NewOrderStore(rdb)
	ctx := context.Background()
	userID := int64(8)
	productID := int64(700)
	key := "cart:" + strconv.FormatInt(userID, 10)

	defer cleanupTestRedis(t, rdb, key)

	// Тест конкурентного доступа
	const goroutines = 10
	done := make(chan bool, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			err := store.AddToCart(ctx, userID, productID)
			assert.NoError(t, err)
			done <- true
		}()
	}

	// Ждем завершения всех горутин
	for i := 0; i < goroutines; i++ {
		<-done
	}

	// Проверяем результат
	cart, err := store.GetCart(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, int64(goroutines), cart[productID])
}

func TestOrderStore_ErrorCases(t *testing.T) {
	rdb := setupTestRedis(t)
	defer rdb.Close()

	store := NewOrderStore(rdb)
	ctx := context.Background()

	// Тестируем обработку ошибок с невалидным контекстом
	canceledCtx, cancel := context.WithCancel(ctx)
	cancel() // Немедленно отменяем контекст

	err := store.AddToCart(canceledCtx, 1, 1)
	assert.Error(t, err)
}
