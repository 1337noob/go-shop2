package orchestrator

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"shop/order_saga/internal/model"
	"shop/order_saga/internal/repository"
	"shop/pkg/command"
	"shop/pkg/event"
	"shop/pkg/outbox"
	"shop/pkg/types"
	"time"

	"github.com/google/uuid"
)

type Orchestrator struct {
	repo   repository.Repository
	outbox outbox.Outbox
}

func NewOrchestrator(repo repository.Repository, outbox outbox.Outbox) *Orchestrator {
	return &Orchestrator{
		repo:   repo,
		outbox: outbox,
	}
}

func (o *Orchestrator) StartSaga(ctx context.Context, s *model.Saga) error {
	log.Println("Starting Saga")

	s.Status = model.StatusInit
	err := o.repo.Create(ctx, s)
	if err != nil {
		return err
	}

	err = o.executeNextStep(ctx, s)
	if err != nil {
		return err
	}

	log.Println("Saga started")
	return nil
}

func (o *Orchestrator) StartCompensating(ctx context.Context, s *model.Saga) error {
	log.Println("Starting Compensating Saga")

	s.Compensating = true
	s.Status = model.StatusCompensating
	s.Steps[s.CurrentStep].CommandStatus = model.StepStatusFailed
	s.CurrentStep--
	err := o.repo.Update(ctx, s)
	if err != nil {
		return err
	}

	return o.compensateNextStep(ctx, s)
}

func (o *Orchestrator) executeNextStep(ctx context.Context, s *model.Saga) error {
	log.Println("Executing NextStep")

	if s.CurrentStep >= len(s.Steps) {
		s.Status = model.StatusCompleted
		err := o.repo.Update(ctx, s)
		if err != nil {
			return err
		}
		return nil
	}

	// TODO save to outbox
	currentStep := s.Steps[s.CurrentStep]
	jsonPayload, err := json.Marshal(s.Payload)
	if err != nil {
		return err
	}
	cmd := command.Command{
		ID:      uuid.New().String(),
		Type:    currentStep.Command,
		SagaID:  s.ID,
		Payload: jsonPayload,
	}
	outboxMessage := outbox.Message{
		ID:        uuid.New().String(),
		Topic:     currentStep.CommandTopic,
		Key:       s.ID,
		Payload:   cmd,
		Status:    outbox.StatusInit,
		CreatedAt: time.Now(),
	}
	err = o.outbox.Publish(ctx, outboxMessage)
	if err != nil {
		return err
	}

	s.Status = model.StatusRunning
	s.Steps[s.CurrentStep].CommandStatus = model.StepStatusRunning
	err = o.repo.Update(ctx, s)
	if err != nil {
		return err
	}

	return nil
}

func (o *Orchestrator) compensateNextStep(ctx context.Context, s *model.Saga) error {
	log.Println("Compensating next step")

	if s.CurrentStep < 0 {
		s.Status = model.StatusCompensated
		return o.repo.Update(ctx, s)
	}

	currentStep := s.Steps[s.CurrentStep]
	if currentStep.Compensate == "" {
		s.Steps[s.CurrentStep].CompensateStatus = model.StepStatusCompleted
		s.CurrentStep--
		err := o.repo.Update(ctx, s)
		if err != nil {
			return err
		}

		err = o.compensateNextStep(ctx, s)
		if err != nil {
			return err
		}

		return nil
	}

	// TODO save to outbox

	s.Steps[s.CurrentStep].CompensateStatus = model.StepStatusRunning
	err := o.repo.Update(ctx, s)
	if err != nil {
		return err
	}

	return nil
}

func (o *Orchestrator) handleSuccessReply(ctx context.Context, s *model.Saga, e event.Event) error {
	log.Println("Handling success event: ", e)

	updatedPayload, err := updatePayload(s.Payload, e)
	if err != nil {
		return err
	}
	s.Payload = updatedPayload
	s.Steps[s.CurrentStep].CommandStatus = model.StepStatusCompleted
	s.CurrentStep++

	err = o.repo.Update(ctx, s)
	if err != nil {
		return err
	}

	err = o.executeNextStep(ctx, s)
	if err != nil {
		return err
	}

	return nil
}

func (o *Orchestrator) handleFailReply(ctx context.Context, s *model.Saga, e event.Event) error {
	log.Println("Handling fail event: ", e)

	if !s.Compensating {
		err := o.StartCompensating(ctx, s)
		if err != nil {
			return err
		}
	} else {
		log.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		log.Println("Saga already compensating")
	}

	return nil
}

func (o *Orchestrator) handleSuccessCompensatingReply(ctx context.Context, s *model.Saga, e event.Event) error {
	log.Println("Handling success compensating event: ", e)

	if s.CurrentStep == 0 && s.Compensating {
		s.Status = model.StatusCompensated
		s.Steps[s.CurrentStep].CompensateStatus = model.StepStatusCompleted
		err := o.repo.Update(ctx, s)
		if err != nil {
			return err
		}

		log.Println("compensating complete")
		return nil
	}

	s.Steps[s.CurrentStep].CompensateStatus = model.StepStatusCompleted
	s.CurrentStep--
	err := o.repo.Update(ctx, s)
	if err != nil {
		return err
	}

	err = o.compensateNextStep(ctx, s)
	if err != nil {
		return err
	}

	return nil
}

func (o *Orchestrator) handleFailCompensatingReply(ctx context.Context, s *model.Saga, e event.Event) error {
	log.Println("Handling fail compensating event: ", e)
	// TODO implement retry
	log.Println("TODO implement retry")

	return nil
}

//func (o *Orchestrator) HandleReply(reply model.Reply) error {
//	log.Println("Handling reply: ", reply)
//
//	s, err := o.repo.Find(reply.SagaId)
//	if err != nil {
//		return err
//	}
//	log.Println("created at: ", s.CreatedAt)
//	log.Println("current command: ", s.Steps[s.CurrentStep].Command)
//	log.Println("current compensation command: ", s.Steps[s.CurrentStep].Compensate)
//
//	if s.Status == model.StatusCompleted || s.Status == model.StatusCompensated {
//		log.Println("Saga is completed or compensated")
//		return nil
//	}
//
//	if reply.Success {
//		log.Println("reply success")
//
//		if s.Compensating {
//			log.Println("Saga is compensating")
//
//			err := o.handleSuccessCompensatingReply(s, reply)
//			if err != nil {
//				return err
//			}
//		} else {
//			log.Println("Saga is not compensating")
//
//			err := o.handleSuccessReply(s, reply)
//			if err != nil {
//				return err
//			}
//		}
//	} else {
//		log.Println("reply failure")
//
//		if s.Compensating {
//			log.Println("Saga is compensating")
//
//			err := o.handleFailCompensatingReply(s, reply)
//			if err != nil {
//				return err
//			}
//		} else {
//			log.Println("Saga is not compensating")
//
//			err := o.handleFailReply(s, reply)
//			if err != nil {
//				return err
//			}
//		}
//	}
//
//	return nil
//}

func (o *Orchestrator) HandleEvent(ctx context.Context, event event.Event) error {
	log.Println("Handling event type: ", event.Type)

	s, err := o.repo.Find(ctx, event.SagaID)
	if err != nil {
		return err
	}

	currentStep := s.Steps[s.CurrentStep]
	switch event.Type {
	case currentStep.CommandSuccessEvent:
		err = o.handleSuccessReply(ctx, s, event)
		if err != nil {
			return err
		}
	case currentStep.CommandFailEvent:
		err = o.handleFailReply(ctx, s, event)
		if err != nil {
			return err
		}
	case currentStep.CompensateSuccessEvent:
		err = o.handleSuccessCompensatingReply(ctx, s, event)
		if err != nil {
			return err
		}
	case currentStep.CompensateFailEvent:
		err = o.handleFailCompensatingReply(ctx, s, event)
		if err != nil {
			return err
		}

	default:
		log.Println("Unknown event type: ", event.Type)
		return errors.New("unknown event type")
	}

	return nil
}

func updatePayload(payload types.SagaPayload, e event.Event) (types.SagaPayload, error) {
	log.Println("Updating payload")
	log.Println(e.Payload)
	switch e.Payload.(type) {

	case event.OrderCreatedPayload:
		payload.OrderId = e.Payload.(event.OrderCreatedPayload).OrderID
		payload.UserID = e.Payload.(event.OrderCreatedPayload).UserID
		payload.PaymentMethodID = e.Payload.(event.OrderCreatedPayload).PaymentMethodID
		payload.OrderItems = e.Payload.(event.OrderCreatedPayload).OrderItems
	case event.OrderCreateFailedPayload:

	case event.ProductsValidatedPayload:
		payload.OrderItems = e.Payload.(event.ProductsValidatedPayload).OrderItems
	case event.ProductsValidationFailedPayload:

	case event.InventoryReservedPayload:
	case event.InventoryReserveFailedPayload:

	case event.PaymentCompletedPayload:
		payload.PaymentID = e.Payload.(event.PaymentCompletedPayload).PaymentID
		payload.PaymentSum = e.Payload.(event.PaymentCompletedPayload).PaymentSum
		payload.PaymentExternalID = e.Payload.(event.PaymentCompletedPayload).PaymentExternalID
	case event.PaymentFailedPayload:

	case event.PaymentRefundedPayload:
	case event.PaymentRefundFailedPayload:

	case event.OrderCompletedPayload:
	case event.OrderCompleteFailedPayload:

	case event.OrderCancelledPayload:

	default:
		log.Println("Unknown payload type: ", e.Payload)
		return types.SagaPayload{}, errors.New("unknown payload type")
	}

	return payload, nil
}

//func mapPayload(cmd command.Command, payload model.SagaPayload) model.SagaPayload {
//	return model.SagaPayload{}
//}
