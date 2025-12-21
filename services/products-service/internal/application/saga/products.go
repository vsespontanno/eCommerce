package saga

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/vsespontanno/eCommerce/services/products-service/internal/presentation/grpc/dto"
	"go.uber.org/zap"
)

type ProductStorage interface {
	ReserveTxn(ctx context.Context, products []*dto.ItemRequest) error
	ReleaseTxn(ctx context.Context, products []*dto.ItemRequest) error
	CommitTxn(ctx context.Context, products []*dto.ItemRequest) error
}

type SagaService struct {
	storage ProductStorage
	logger  *zap.SugaredLogger
}

func NewSagaService(storage ProductStorage, logger *zap.SugaredLogger) *SagaService {
	return &SagaService{storage: storage, logger: logger}
}

func (s *SagaService) Reserve(ctx context.Context, products []*dto.ItemRequest) error {
	s.logger.Infow("Reserving products in saga", "products ", products)
	return s.execWithRetry(ctx, "reserve", func() error {
		return s.storage.ReserveTxn(ctx, products)
	})
}

func (s *SagaService) Release(ctx context.Context, products []*dto.ItemRequest) error {
	s.logger.Infow("Releasing products in saga", "products ", products)
	return s.execWithRetry(ctx, "release", func() error {
		return s.storage.ReleaseTxn(ctx, products)
	})
}

func (s *SagaService) Commit(ctx context.Context, products []*dto.ItemRequest) error {
	s.logger.Infow("Committing products in saga", "products ", products)
	return s.execWithRetry(ctx, "commit", func() error {
		return s.storage.CommitTxn(ctx, products)
	})
}

// execWithRetry — обёртка для любых транзакций, защищает от transient ошибок (deadlock, serialization failure).
func (s *SagaService) execWithRetry(ctx context.Context, op string, fn func() error) error {
	const maxAttempts = 5
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}
		if IsTransientErr(err) {
			backoff := time.Duration(attempt*attempt) * 50 * time.Millisecond
			time.Sleep(backoff)
			continue
		}
		return fmt.Errorf("%s failed: %w", op, err)
	}
	return fmt.Errorf("%s failed after %d attempts", op, maxAttempts)
}

// IsTransientErr — распознаёт временные ошибки БД, на которые стоит делать retry.
func IsTransientErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "deadlock detected") ||
		strings.Contains(msg, "could not serialize access") ||
		strings.Contains(msg, "serialization failure") ||
		strings.Contains(msg, "retry transaction")
}
