package saga

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/vsespontanno/eCommerce/products-service/internal/grpc/dto"
)

type ProductStorage interface {
	ReserveTxn(ctx context.Context, products []*dto.ItemRequest) error
	ReleaseProducts(ctx context.Context, products []*dto.ItemRequest) error
	CommitProducts(ctx context.Context, products []*dto.ItemRequest) error
}

type SagaService struct {
	storage ProductStorage
}

func NewSagaService(storage ProductStorage) *SagaService {
	return &SagaService{
		storage: storage,
	}
}

// ReserveProducts пытается зарезервировать набор товаров в одной транзакции.
// items: список (productID, qty)
// Возвращает nil если всё ок, или ErrNotEnoughStock + детали, или другую ошибку.
func (s *SagaService) ReserveProducts(ctx context.Context, products []*dto.ItemRequest) error {
	const maxAttempts = 5
	var attempt int
	for attempt = 1; attempt <= maxAttempts; attempt++ {
		err := s.storage.ReserveTxn(ctx, products)
		if err == nil {
			return nil
		}
		// Если transient (deadlock/serialization), ретраим с backoff
		if IsTransientErr(err) {
			// экспоненциальный бэк-офф
			backoff := time.Duration(attempt*attempt) * 50 * time.Millisecond
			time.Sleep(backoff)
			continue
		}
		return err
	}
	return fmt.Errorf("reserve failed after %d attempts", maxAttempts)
}

// isTransientErr пытается распознать transient ошибки БД, на которые стоит делать retry.
// Здесь простая реализация: проверяем текст ошибки на известные фразы. потенциальный туду, можно улучшать
func IsTransientErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "deadlock detected") || strings.Contains(msg, "could not serialize access") ||
		strings.Contains(msg, "serialization failure") || strings.Contains(msg, "retry transaction") {
		return true
	}
	return false
}
