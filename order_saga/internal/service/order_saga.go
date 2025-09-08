package service

import (
	"context"
	"log"
	"shop/order_saga/internal/model"
	"shop/order_saga/internal/orchestrator"
	"shop/pkg/types"
)

type OrderSagaService struct {
	orchestrator *orchestrator.Orchestrator
	logger       *log.Logger
}

func NewOrderSagaService(orc *orchestrator.Orchestrator, logger *log.Logger) *OrderSagaService {
	return &OrderSagaService{
		orchestrator: orc,
		logger:       logger,
	}
}

func (s *OrderSagaService) Create(ctx context.Context, userID string, items []types.Item, paymentMethod string) error {
	s.logger.Printf("Create order saga start")

	saga := model.NewCreateOrderSaga(userID, items, paymentMethod)
	err := s.orchestrator.StartSaga(ctx, saga)
	if err != nil {
		s.logger.Printf("Create order saga failed: %v", err)
		return err
	}

	return nil
}
