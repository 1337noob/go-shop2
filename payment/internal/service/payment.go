package service

import (
	"context"
	"encoding/json"
	"log"
	"shop/payment/internal/model"
	"shop/payment/internal/repository"
	"shop/pkg/event"

	"github.com/google/uuid"
)

type PaymentService struct {
	paymentRepo repository.PaymentRepository
	methodRepo  repository.MethodRepository
	logger      *log.Logger
}

func NewPaymentService(paymentRepo repository.PaymentRepository, methodRepo repository.MethodRepository, logger *log.Logger) *PaymentService {
	return &PaymentService{
		paymentRepo: paymentRepo,
		methodRepo:  methodRepo,
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

	method, err := s.methodRepo.FindByID(ctx, pay.MethodID)
	if err != nil {
		s.logger.Printf("Find Payment Method Failed: %v", err)
		return model.Payment{}, e, err
	}

	// fake charge delay
	//time.Sleep(time.Second * 2)

	fakeExternalID := uuid.New().String()
	eventID := uuid.New().String()
	completedStatus := model.PaymentStatusCompleted
	failedStatus := model.PaymentStatusFailed

	if pay.MethodID == "method-fail" {
		err = s.paymentRepo.UpdateStatus(ctx, pay.ID, failedStatus)
		if err != nil {
			s.logger.Printf("Update Payment Failed: %v", err)
			return model.Payment{}, e, err
		}
		pay.Status = failedStatus
		p := event.PaymentCompletedPayload{
			PaymentID:         pay.ID,
			OrderID:           pay.OrderID,
			UserID:            pay.UserID,
			PaymentSum:        pay.Amount,
			PaymentMethodID:   pay.MethodID,
			PaymentExternalID: fakeExternalID,
			PaymentStatus:     string(pay.Status),
		}
		jsonPayload, err := json.Marshal(p)
		if err != nil {
			s.logger.Println("failed to marshal payload", "error", err)
			return model.Payment{}, e, err
		}
		e = event.Event{
			ID:      eventID,
			Type:    event.PaymentFailed,
			Payload: jsonPayload,
		}

		return pay, e, nil
	}

	err = s.paymentRepo.UpdateStatus(ctx, pay.ID, completedStatus)
	if err != nil {
		s.logger.Printf("Update Payment Failed: %v", err)
		return model.Payment{}, e, err
	}
	pay.Status = completedStatus
	p := event.PaymentCompletedPayload{
		PaymentID:         pay.ID,
		OrderID:           pay.OrderID,
		UserID:            pay.UserID,
		PaymentSum:        pay.Amount,
		PaymentMethodID:   pay.MethodID,
		PaymentExternalID: fakeExternalID,
		PaymentType:       method.PaymentType,
		PaymentGateway:    method.Gateway,
		PaymentStatus:     string(pay.Status),
	}
	jsonPayload, err := json.Marshal(p)
	if err != nil {
		s.logger.Println("failed to marshal payload", "error", err)
		return model.Payment{}, e, err
	}
	e = event.Event{
		ID:      eventID,
		Type:    event.PaymentCompleted,
		Payload: jsonPayload,
	}

	return pay, e, nil
}
