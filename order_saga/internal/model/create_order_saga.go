package model

import (
	"shop/pkg/command"
	"shop/pkg/event"
	"shop/pkg/types"
	"time"

	"github.com/google/uuid"
)

func NewCreateOrderSaga(userID string, items []types.Item) *Saga {
	steps := []Step{
		// create order
		{
			Command:                command.CreateOrder,
			CommandStatus:          StepStatusInit,
			CommandSuccessEvent:    event.OrderCreated,
			CommandFailEvent:       event.OrderCreateFailed,
			Compensate:             command.CancelOrder,
			CompensateStatus:       StepStatusInit,
			CompensateSuccessEvent: event.OrderCancelled,
			CompensateFailEvent:    event.OrderCancelFailed,
			CommandTopic:           "order-commands",
		},
		// validate products
		{
			Command:                command.ValidateProducts,
			CommandStatus:          StepStatusInit,
			CommandSuccessEvent:    event.ProductsValidated,
			CommandFailEvent:       event.ProductsValidationFailed,
			Compensate:             "",
			CompensateStatus:       "",
			CompensateSuccessEvent: "",
			CompensateFailEvent:    "",
			CommandTopic:           "order-commands",
		},
		// reserve inventory
		{
			Command:                command.ReserveInventory,
			CommandStatus:          StepStatusInit,
			CommandSuccessEvent:    event.InventoryReserved,
			CommandFailEvent:       event.InventoryReserveFailed,
			Compensate:             command.ReleaseInventory,
			CompensateStatus:       StepStatusInit,
			CompensateSuccessEvent: event.InventoryReleased,
			CompensateFailEvent:    event.InventoryReleaseFailed,
			CommandTopic:           "inventory-commands",
		},
		// process payment
		{
			Command:                command.ProcessPayment,
			CommandStatus:          StepStatusInit,
			CommandSuccessEvent:    event.PaymentCompleted,
			CommandFailEvent:       event.PaymentFailed,
			Compensate:             command.RefundPayment,
			CompensateStatus:       StepStatusInit,
			CompensateSuccessEvent: event.PaymentRefunded,
			CompensateFailEvent:    event.PaymentRefundFailed,
			CommandTopic:           "payment-commands",
		},
		// complete order
		{
			Command:                command.CompleteOrder,
			CommandStatus:          StepStatusInit,
			CommandSuccessEvent:    event.OrderCompleted,
			CommandFailEvent:       event.OrderCompleteFailed,
			Compensate:             "",
			CompensateStatus:       "",
			CompensateSuccessEvent: "",
			CompensateFailEvent:    "",
			CommandTopic:           "order-commands",
		},
	}

	return &Saga{
		ID:          uuid.New().String(),
		CurrentStep: 0,
		Status:      StatusInit,
		Steps:       steps,
		Payload: types.SagaPayload{
			UserID:     userID,
			OrderItems: items,
		},
		Compensating: false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}
