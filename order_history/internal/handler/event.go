package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"shop/order_history/internal/model"
	"shop/order_history/internal/repository"
	"shop/pkg/broker"
	"shop/pkg/event"
	"shop/pkg/inbox"
	"time"
)

type EventHandler struct {
	db        *sql.DB
	inbox     inbox.Inbox
	orderRepo repository.OrderRepository
	logger    *log.Logger
}

func NewEventHandler(db *sql.DB, inbox inbox.Inbox, orderRepo repository.OrderRepository, logger *log.Logger) *EventHandler {
	return &EventHandler{
		db:        db,
		inbox:     inbox,
		orderRepo: orderRepo,
		logger:    logger,
	}
}

func (h *EventHandler) Handle(message broker.Message) error {
	h.logger.Printf("Handling message %s from %s", message.Key, message.Topic)
	h.logger.Println(string(message.Value))

	var e event.Event
	err := json.Unmarshal(message.Value, &e)
	if err != nil {
		h.logger.Printf("Error unmarshalling command: %s", err)
	}
	h.logger.Printf("Handling event: %+v", e)

	tx, err := h.db.Begin()
	if err != nil {
		h.logger.Println("Failed to begin transaction", "error", err)
		return err
	}
	defer tx.Rollback()

	ctxWithTx := context.WithValue(context.Background(), "tx", tx)

	exists, err := h.inbox.Exists(ctxWithTx, e.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		h.logger.Printf("Failed to check if the message %s exists: %s", e.ID, err)
		return err
	}

	if exists {
		h.logger.Println("Ignore existing message")
		return nil
	} else {
		h.logger.Println("Message not exists")
		// store to inbox
		inboxMessage := inbox.Message{
			MessageID:   e.ID,
			MessageType: string(e.Type),
			Topic:       message.Topic,
			Key:         message.Key,
			Payload:     message.Value,
			Status:      inbox.StatusPending,
			CreatedAt:   time.Now(),
		}
		err = h.inbox.Store(ctxWithTx, inboxMessage)
		if err != nil {
			h.logger.Printf("Error storing inbox message: %s", err)
			return err
		}
		h.logger.Printf("Successfully stored inbox message: %+v", inboxMessage)
	}

	err = tx.Commit()
	if err != nil {
		h.logger.Println("Failed to commit transaction", "error", err)
		return err
	}

	tx, err = h.db.Begin()
	if err != nil {
		h.logger.Println("Failed to begin transaction", "error", err)
		return err
	}
	defer tx.Rollback()

	ctxWithTx = context.WithValue(context.Background(), "tx", tx)

	switch e.Type {
	case event.OrderCreated:
		err = h.handleOrderCreated(ctxWithTx, e)
		if err != nil {
			h.logger.Printf("Error handling order created: %s", err)
			return err
		}
	case event.ProductsValidated:
		err = h.handleProductsValidated(ctxWithTx, e)
		if err != nil {
			h.logger.Printf("Error handling products validated: %s", err)
			return err
		}
	case event.InventoryReserved:
		err = h.handleInventoryReserved(ctxWithTx, e)
		if err != nil {
			h.logger.Printf("Error handling inventory reserved: %s", err)
			return err
		}
	case event.PaymentCompleted:
		err = h.handlePaymentCompleted(ctxWithTx, e)
		if err != nil {
			h.logger.Printf("Error handling payment completed: %s", err)
			return err
		}
	case event.OrderCompleted:
		err = h.handleOrderCompleted(ctxWithTx, e)
		if err != nil {
			h.logger.Printf("Error handling order completed: %s", err)
			return err
		}
	case event.InventoryReserveFailed:
		err = h.handleInventoryReserveFailed(ctxWithTx, e)
		if err != nil {
			h.logger.Printf("Error handling inventory reserve failed: %s", err)
			return err
		}
	case event.PaymentFailed:
		err = h.handlePaymentFailed(ctxWithTx, e)
		if err != nil {
			h.logger.Printf("Error handling payment failed: %s", err)
			return err
		}
	default:
		h.logger.Printf("Invalid event type: %s", e.Type)
	}

	err = h.inbox.MarkAsCompleted(ctxWithTx, e.ID)
	if err != nil {
		h.logger.Printf("failed to mark as completed: %s", err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		h.logger.Println("failed to commit transaction", "error", err)
		return err
	}

	return nil
}

func (h *EventHandler) handleOrderCreated(ctx context.Context, e event.Event) error {
	h.logger.Printf("Handling order created event: %+v", e)

	var payload event.OrderCreatedPayload
	err := json.Unmarshal(e.Payload, &payload)
	if err != nil {
		h.logger.Printf("Error unmarshalling payload: %s", err)
		return err
	}

	order := &model.Order{
		ID:              payload.OrderID,
		UserID:          payload.UserID,
		PaymentMethodID: payload.PaymentMethodID,
		OrderItems:      payload.OrderItems,
		CreatedAt:       payload.CreatedAt,
		UpdatedAt:       time.Now(),
		Status:          model.StatusOrderCreated,
	}

	err = h.orderRepo.Create(ctx, order)
	if err != nil {
		h.logger.Printf("Error creating order: %s", err)
		return err
	}

	return nil
}

func (h *EventHandler) handleProductsValidated(ctx context.Context, e event.Event) error {
	h.logger.Printf("Handling products validated event: %+v", e)

	var payload event.ProductsValidatedPayload
	err := json.Unmarshal(e.Payload, &payload)
	if err != nil {
		h.logger.Printf("Error unmarshalling payload: %s", err)
		return err
	}

	order, err := h.orderRepo.FindByID(ctx, payload.OrderID)
	if err != nil {
		h.logger.Printf("Error finding order: %s", err)
		return err
	}

	for i, orderItem := range order.OrderItems {
		for _, payloadOrderItem := range payload.OrderItems {
			if payloadOrderItem.ProductID == orderItem.ProductID {
				order.OrderItems[i].Name = payloadOrderItem.Name
				order.OrderItems[i].Price = payloadOrderItem.Price
			}

		}
	}
	order.Status = model.StatusProductsValidated
	order.UpdatedAt = time.Now()

	err = h.orderRepo.Update(ctx, order)
	if err != nil {
		h.logger.Printf("Error updating order: %s", err)
		return err
	}

	return nil
}

func (h *EventHandler) handleInventoryReserved(ctx context.Context, e event.Event) error {
	h.logger.Printf("Handling inventory reserved event: %+v", e)

	var payload event.InventoryReservedPayload
	err := json.Unmarshal(e.Payload, &payload)
	if err != nil {
		h.logger.Printf("Error unmarshalling payload: %s", err)
		return err
	}

	order, err := h.orderRepo.FindByID(ctx, payload.OrderID)
	if err != nil {
		h.logger.Printf("Error finding order: %s", err)
		return err
	}

	order.Status = model.StatusInventoryReserved
	order.UpdatedAt = time.Now()

	err = h.orderRepo.Update(ctx, order)
	if err != nil {
		h.logger.Printf("Error updating order: %s", err)
		return err
	}

	return nil
}

func (h *EventHandler) handlePaymentCompleted(ctx context.Context, e event.Event) error {
	h.logger.Printf("Handling payment completed event: %+v", e)

	var payload event.PaymentCompletedPayload
	err := json.Unmarshal(e.Payload, &payload)
	if err != nil {
		h.logger.Printf("Error unmarshalling payload: %s", err)
		return err
	}

	order, err := h.orderRepo.FindByID(ctx, payload.OrderID)
	if err != nil {
		h.logger.Printf("Error finding order: %s", err)
		return err
	}

	order.PaymentID = payload.PaymentID
	order.PaymentExternalID = payload.PaymentExternalID
	order.PaymentSum = payload.PaymentSum
	order.PaymentType = payload.PaymentType
	order.PaymentGateway = payload.PaymentGateway
	order.PaymentStatus = payload.PaymentStatus
	order.Status = model.StatusPaymentCompleted
	order.UpdatedAt = time.Now()

	err = h.orderRepo.Update(ctx, order)
	if err != nil {
		h.logger.Printf("Error updating order: %s", err)
		return err
	}

	return nil
}

func (h *EventHandler) handleOrderCompleted(ctx context.Context, e event.Event) error {
	h.logger.Printf("Handling order completed event: %+v", e)

	var payload event.OrderCompletedPayload
	err := json.Unmarshal(e.Payload, &payload)
	if err != nil {
		h.logger.Printf("Error unmarshalling payload: %s", err)
		return err
	}

	order, err := h.orderRepo.FindByID(ctx, payload.OrderID)
	if err != nil {
		h.logger.Printf("Error finding order: %s", err)
		return err
	}

	order.Status = model.StatusOrderCompleted
	order.UpdatedAt = time.Now()

	err = h.orderRepo.Update(ctx, order)
	if err != nil {
		h.logger.Printf("Error updating order: %s", err)
		return err
	}

	return nil
}

func (h *EventHandler) handleInventoryReserveFailed(ctx context.Context, e event.Event) error {
	h.logger.Printf("Handling inventory reserve failed event: %+v", e)

	var payload event.InventoryReserveFailedPayload
	err := json.Unmarshal(e.Payload, &payload)
	if err != nil {
		h.logger.Printf("Error unmarshalling payload: %s", err)
		return err
	}

	order, err := h.orderRepo.FindByID(ctx, payload.OrderID)
	if err != nil {
		h.logger.Printf("Error finding order: %s", err)
		return err
	}

	order.Status = model.StatusInventoryReserveFailed
	order.UpdatedAt = time.Now()

	err = h.orderRepo.Update(ctx, order)
	if err != nil {
		h.logger.Printf("Error updating order: %s", err)
		return err
	}

	return nil
}

func (h *EventHandler) handlePaymentFailed(ctx context.Context, e event.Event) error {
	h.logger.Printf("Handling payment failed event: %+v", e)

	var payload event.PaymentFailedPayload
	err := json.Unmarshal(e.Payload, &payload)
	if err != nil {
		h.logger.Printf("Error unmarshalling payload: %s", err)
		return err
	}

	order, err := h.orderRepo.FindByID(ctx, payload.OrderID)
	if err != nil {
		h.logger.Printf("Error finding order: %s", err)
		return err
	}

	order.Status = model.StatusPaymentFailed
	order.UpdatedAt = time.Now()

	err = h.orderRepo.Update(ctx, order)
	if err != nil {
		h.logger.Printf("Error updating order: %s", err)
		return err
	}

	return nil
}
