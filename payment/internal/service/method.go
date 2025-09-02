package service

import (
	"context"
	"github.com/google/uuid"
	"log"
	"shop/payment/internal/model"
	"shop/payment/internal/repository"
	"shop/pkg/event"
)

type MethodService struct {
	methodRepo repository.MethodRepository
	logger     *log.Logger
}

func NewMethodService(repo repository.MethodRepository, logger *log.Logger) *MethodService {
	return &MethodService{methodRepo: repo, logger: logger}
}

func (s *MethodService) Store(ctx context.Context, method model.Method) (model.Method, event.Event, error) {
	var e event.Event

	method, err := s.methodRepo.Create(ctx, method)
	if err != nil {
		s.logger.Println("failed to create method", "error", err)
		return model.Method{}, e, err
	}

	p := event.PaymentMethodCreatedPayload{
		ID:          method.ID,
		UserID:      method.UserID,
		Gateway:     method.Gateway,
		PaymentType: method.PaymentType,
		Token:       method.Token,
	}
	e = event.Event{
		ID:      uuid.New().String(),
		Type:    event.TypePaymentMethodCreated,
		Payload: p,
	}

	return method, e, nil
}
