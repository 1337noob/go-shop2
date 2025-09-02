package service

import (
	"context"
	"github.com/google/uuid"
	"log"
	"shop/payment/internal/model"
	"shop/payment/internal/repository"
	"shop/pkg/event"
	"time"
)

type PaymentService struct {
	paymentRepo repository.PaymentRepository
	logger      *log.Logger
}

func NewPaymentService(paymentRepo repository.PaymentRepository, logger *log.Logger) *PaymentService {
	return &PaymentService{
		paymentRepo: paymentRepo,
		logger:      logger,
	}
}

func (s *PaymentService) Process(ctx context.Context, payment model.Payment) (model.Payment, event.Event, error) {
	var e event.Event

	pay, err := s.paymentRepo.Create(ctx, payment)
	if err != nil {
		s.logger.Printf("Create Payment Failed: %v", err)
		return model.Payment{}, e, err
	}

	// fake charge delay
	time.Sleep(time.Second * 2)

	eventID := uuid.New().String()
	completedStatus := model.PaymentStatusCompleted
	failedStatus := model.PaymentStatusFailed

	if pay.MethodID == "fail" {
		err = s.paymentRepo.UpdateStatus(ctx, pay.ID, failedStatus)
		if err != nil {
			s.logger.Printf("Update Payment Failed: %v", err)
			return model.Payment{}, e, err
		}
		pay.Status = failedStatus
		p := event.PaymentCompletedPayload{
			PaymentID:  pay.ID,
			OrderID:    pay.OrderID,
			UserID:     pay.UserID,
			Amount:     pay.Amount,
			MethodID:   pay.MethodID,
			ExternalID: pay.ExternalID,
			Status:     string(pay.Status),
		}
		e = event.Event{
			ID:      eventID,
			Type:    event.TypePaymentFailed,
			Payload: p,
		}

		return pay, e, nil
	}

	pay.Status = completedStatus
	p := event.PaymentCompletedPayload{
		PaymentID:  pay.ID,
		OrderID:    pay.OrderID,
		UserID:     pay.UserID,
		Amount:     pay.Amount,
		MethodID:   pay.MethodID,
		ExternalID: pay.ExternalID,
		Status:     string(pay.Status),
	}
	e = event.Event{
		ID:      eventID,
		Type:    event.TypePaymentCompleted,
		Payload: p,
	}

	return pay, e, nil
}
