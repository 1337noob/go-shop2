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
	logger *log.Logger
}

func NewOrchestrator(repo repository.Repository, outbox outbox.Outbox, logger *log.Logger) *Orchestrator {
	return &Orchestrator{
		repo:   repo,
		outbox: outbox,
		logger: logger,
	}
}

func (o *Orchestrator) StartSaga(ctx context.Context, s *model.Saga) error {
	o.logger.Println("Saga start")

	s.Status = model.StatusInit
	err := o.repo.Create(ctx, s)
	if err != nil {
		return err
	}

	err = o.executeNextStep(ctx, s)
	if err != nil {
		return err
	}

	o.logger.Println("Saga started")
	return nil
}

func (o *Orchestrator) StartCompensating(ctx context.Context, s *model.Saga) error {
	o.logger.Println("Saga start compensating")

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
	o.logger.Println("Saga start execute next step")

	if s.CurrentStep >= len(s.Steps) {
		s.Status = model.StatusCompleted
		err := o.repo.Update(ctx, s)
		if err != nil {
			return err
		}
		return nil
	}

	currentStep := s.Steps[s.CurrentStep]
	jsonPayload, err := o.mapToJsonPayload(currentStep.Command, s.Payload)
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

	o.logger.Println("Saga finish next step")
	return nil
}

func (o *Orchestrator) compensateNextStep(ctx context.Context, s *model.Saga) error {
	o.logger.Println("Saga start compensate next step")

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

	jsonPayload, err := o.mapToJsonPayload(currentStep.Compensate, s.Payload)
	if err != nil {
		return err
	}
	cmd := command.Command{
		ID:      uuid.New().String(),
		Type:    currentStep.Compensate,
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

	s.Steps[s.CurrentStep].CompensateStatus = model.StepStatusRunning
	err = o.repo.Update(ctx, s)
	if err != nil {
		return err
	}

	o.logger.Println("Saga finish compensate next step")
	return nil
}

func (o *Orchestrator) handleSuccessReply(ctx context.Context, s *model.Saga, e event.Event) error {
	o.logger.Println("Saga start handle success event: ", e)

	updatedPayload, err := o.updatePayload(s.Payload, e)
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

	o.logger.Println("Saga finish handle success event: ", e)
	return nil
}

func (o *Orchestrator) handleFailReply(ctx context.Context, s *model.Saga, e event.Event) error {
	o.logger.Println("Saga start handle fail event: ", e)

	if !s.Compensating {
		err := o.StartCompensating(ctx, s)
		if err != nil {
			return err
		}
	} else {
		o.logger.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		o.logger.Println("Saga already compensating")
	}

	o.logger.Println("Saga finish handle fail event: ", e)
	return nil
}

func (o *Orchestrator) handleSuccessCompensatingReply(ctx context.Context, s *model.Saga, e event.Event) error {
	o.logger.Println("Saga start handle success compensating event: ", e)

	if s.CurrentStep == 0 && s.Compensating {
		s.Status = model.StatusCompensated
		s.Steps[s.CurrentStep].CompensateStatus = model.StepStatusCompleted
		err := o.repo.Update(ctx, s)
		if err != nil {
			return err
		}

		o.logger.Println("Saga compensating complete")
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

	o.logger.Println("Saga finish handle success compensating event: ", e)
	return nil
}

func (o *Orchestrator) handleFailCompensatingReply(ctx context.Context, s *model.Saga, e event.Event) error {
	o.logger.Println("Saga start handle fail compensating event: ", e)
	// TODO implement retry
	log.Println("TODO implement retry ?")

	o.logger.Println("Saga finish handle fail compensating event: ", e)
	return nil
}

func (o *Orchestrator) HandleEvent(ctx context.Context, event event.Event) error {
	o.logger.Println("Saga start handle event type: ", event.Type)

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

	o.logger.Println("Saga finish handle event type: ", event.Type)
	return nil
}

func (o *Orchestrator) updatePayload(sagaPayload types.SagaPayload, e event.Event) (types.SagaPayload, error) {
	o.logger.Println("Saga start update payload")
	o.logger.Println(e.Payload)

	switch e.Type {

	case event.OrderCreated:
		var eventPayload event.OrderCreatedPayload
		err := json.Unmarshal(e.Payload, &eventPayload)
		if err != nil {
			return types.SagaPayload{}, err
		}
		sagaPayload.OrderID = eventPayload.OrderID
		//sagaPayload.UserID = eventPayload.UserID
		//sagaPayload.PaymentMethodID = eventPayload.PaymentMethodID
		//sagaPayload.OrderItems = eventPayload.OrderItems
	case event.OrderCreateFailed:

	case event.ProductsValidated:
		var eventPayload event.ProductsValidatedPayload
		err := json.Unmarshal(e.Payload, &eventPayload)
		if err != nil {
			return types.SagaPayload{}, err
		}
		for i, item := range sagaPayload.OrderItems {
			for _, eItem := range eventPayload.OrderItems {
				if item.ProductID == eItem.ProductID {
					o.logger.Println("Updating order item #", i)
					o.logger.Println("item: ", item)
					o.logger.Println("eItem: ", eItem)
					sagaPayload.OrderItems[i].Price = eItem.Price
					sagaPayload.OrderItems[i].Name = eItem.Name
					//sagaPayload.OrderItems[i].Quantity = item.Quantity
					break
				}
			}
		}
	case event.ProductsValidationFailed:

	case event.InventoryReserved:
	case event.InventoryReserveFailed:

	case event.PaymentCompleted:
		var eventPayload event.PaymentCompletedPayload
		err := json.Unmarshal(e.Payload, &eventPayload)
		if err != nil {
			return types.SagaPayload{}, err
		}
		sagaPayload.PaymentID = eventPayload.PaymentID
		sagaPayload.PaymentSum = eventPayload.PaymentSum
		sagaPayload.PaymentExternalID = eventPayload.PaymentExternalID
	case event.PaymentFailed:

	case event.PaymentRefunded:
	case event.PaymentRefundFailed:

	case event.OrderCompleted:
	case event.OrderCompleteFailed:

	case event.OrderCancelled:

	default:
		log.Println("Unknown payload type: ", e.Payload)
		return types.SagaPayload{}, errors.New("unknown payload type")
	}

	o.logger.Println("Saga finish update payload")
	return sagaPayload, nil
}

func (o *Orchestrator) mapToJsonPayload(commandType command.Type, payload types.SagaPayload) ([]byte, error) {
	o.logger.Println("Saga start map payload")

	var result []byte
	var err error

	switch commandType {
	case command.CreateOrder:
		newPayload := command.CreateOrderPayload{
			UserID:          payload.UserID,
			PaymentMethodID: payload.PaymentMethodID,
			OrderItems:      payload.OrderItems,
		}
		result, err = json.Marshal(newPayload)
		if err != nil {
			return nil, err
		}
	case command.ValidateProducts:
		newPayload := command.ValidateProductsPayload{
			OrderItems: payload.OrderItems,
		}
		result, err = json.Marshal(newPayload)
		if err != nil {
			return nil, err
		}
	case command.ReserveInventory:
		newPayload := command.ReserveInventoryPayload{
			OrderItems: payload.OrderItems,
		}
		result, err = json.Marshal(newPayload)
		if err != nil {
			return nil, err
		}
	case command.ProcessPayment:
		newPayload := command.ProcessPaymentPayload{
			OrderID:         payload.OrderID,
			UserID:          payload.UserID,
			PaymentSum:      payload.PaymentSum,
			PaymentMethodID: payload.PaymentMethodID,
		}
		result, err = json.Marshal(newPayload)
		if err != nil {
			return nil, err
		}
	case command.CompleteOrder:
		newPayload := command.CompleteOrderPayload{
			OrderID: payload.OrderID,
		}
		result, err = json.Marshal(newPayload)
		if err != nil {
			return nil, err
		}
	case command.ReleaseInventory:
		newPayload := command.ReleaseInventoryPayload{
			OrderItems: payload.OrderItems,
		}
		result, err = json.Marshal(newPayload)
		if err != nil {
			return nil, err
		}
	case command.RefundPayment:
		newPayload := command.RefundPaymentPayload{
			PaymentID: payload.PaymentID,
		}
		result, err = json.Marshal(newPayload)
		if err != nil {
			return nil, err
		}
	case command.CancelOrder:
		newPayload := command.CancelOrderPayload{
			OrderID: payload.OrderID,
		}
		result, err = json.Marshal(newPayload)
		if err != nil {
			return nil, err
		}

	default:
		log.Println("Unknown command type: ", commandType)
		return result, errors.New("unknown command type")
	}

	o.logger.Println("Saga finish map payload")
	return result, nil
}
