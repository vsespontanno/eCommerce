package saga

import (
	"github.com/vsespontanno/eCommerce/order-service/internal/config"
	"github.com/vsespontanno/eCommerce/order-service/internal/domain/saga/interfaces"
)

type Orchestrator struct {
	sagaRepo interfaces.SagaRepo
	config   *config.Config
}
