package model

import (
	"shop/pkg/command"
	"shop/pkg/event"
	"time"

	"github.com/google/uuid"
)

func NewCreateOrderSaga(userID string, items []OrderItem) *Saga {
	steps := []Step{
		// create order
		{
			Command:                command.CreateOrder,
			CommandStatus:          StepStatusInit,
			CommandSuccessEvent:    event.TypeOrderCreated,
			CommandFailEvent:       event.TypeOrderCreateFailed,
			Compensate:             command.CancelOrder,
			CompensateStatus:       StepStatusInit,
			CompensateSuccessEvent: event.TypeOrderCancelled,
			CompensateFailEvent:    event.TypeOrderCancelFailed,
			CommandTopic:           "order-commands",
		},
		// validate products
		{
			Command:                command.ValidateProducts,
			CommandStatus:          StepStatusInit,
			CommandSuccessEvent:    event.TypeProductsValidated,
			CommandFailEvent:       event.TypeProductsValidationFailed,
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
			CommandSuccessEvent:    event.TypeInventoryReserved,
			CommandFailEvent:       event.TypeInventoryReserveFailed,
			Compensate:             command.ReleaseInventory,
			CompensateStatus:       StepStatusInit,
			CompensateSuccessEvent: event.TypeInventoryReleased,
			CompensateFailEvent:    event.TypeInventoryReleaseFailed,
			CommandTopic:           "inventory-commands",
		},
		// process payment
		{
			Command:                command.ProcessPayment,
			CommandStatus:          StepStatusInit,
			CommandSuccessEvent:    event.TypePaymentCompleted,
			CommandFailEvent:       event.TypePaymentFailed,
			Compensate:             command.RefundPayment,
			CompensateStatus:       StepStatusInit,
			CompensateSuccessEvent: event.TypePaymentRefunded,
			CompensateFailEvent:    event.TypePaymentRefundFailed,
			CommandTopic:           "payment-commands",
		},
		// complete order
		{
			Command:                command.CompleteOrder,
			CommandStatus:          StepStatusInit,
			CommandSuccessEvent:    event.TypeOrderCompleted,
			CommandFailEvent:       event.TypeOrderCompleteFailed,
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
		Payload: SagaPayload{
			UserID:     userID,
			OrderItems: items,
		},
		Compensating: false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}
