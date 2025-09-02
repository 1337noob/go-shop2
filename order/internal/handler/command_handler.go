package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"shop/order/internal/model"
	"shop/order/internal/service"
	"shop/pkg/broker"
	"shop/pkg/command"
	"shop/pkg/event"
	"shop/pkg/inbox"
	"shop/pkg/outbox"
	"time"

	"github.com/google/uuid"
)

type CommandHandler struct {
	db           *sql.DB
	orderService *service.OrderService
	inbox        inbox.Inbox
	outbox       outbox.Outbox
	logger       *log.Logger
}

func NewCommandHandler(db *sql.DB, orderService *service.OrderService, inbox inbox.Inbox, outbox outbox.Outbox, logger *log.Logger) *CommandHandler {
	return &CommandHandler{
		db:           db,
		orderService: orderService,
		inbox:        inbox,
		outbox:       outbox,
		logger:       logger,
	}
}

func (h *CommandHandler) Handle(message broker.Message) error {
	h.logger.Printf("Handling message %s from %s", message.Key, message.Topic)

	var cmd command.Command
	err := json.Unmarshal(message.Value, &cmd)
	if err != nil {
		h.logger.Printf("Error unmarshalling command: %s", err)
	}
	h.logger.Printf("Handling command: %+v", cmd)

	tx, err := h.db.Begin()
	if err != nil {
		h.logger.Println("Failed to begin transaction", "error", err)
		return err
	}
	defer tx.Rollback()

	ctxWithTx := context.WithValue(context.Background(), "tx", tx)

	exists, err := h.inbox.Exists(ctxWithTx, cmd.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		h.logger.Printf("Failed to check if the message %s exists: %s", cmd.ID, err)
		return err
	}

	if exists {
		h.logger.Println("Ignore existing message")
		return nil
	} else {
		h.logger.Println("Message not exists")
		// store to inbox
		inboxMessage := inbox.Message{
			MessageID:   cmd.ID,
			MessageType: string(cmd.Type),
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

	var e event.Event

	switch cmd.Type {
	case command.CreateOrder:
		h.logger.Printf("Create order command: %+v", cmd)
		e, err = h.handleCreateOrder(ctxWithTx, cmd.Payload)
		if err != nil {
			return err
		}
	case command.CompleteOrder:
		h.logger.Printf("Complete order command: %+v", cmd)
		e, err = h.handleCompleteOrder(ctxWithTx, cmd.Payload)
		if err != nil {
			return err
		}
	default:
		return errors.New("invalid command")
	}

	e.SagaID = cmd.SagaID

	outboxMessage := outbox.Message{
		ID:        uuid.New().String(),
		Topic:     "order-events",
		Key:       cmd.SagaID,
		Payload:   e,
		Status:    outbox.StatusInit,
		CreatedAt: time.Now(),
	}
	err = h.outbox.Publish(ctxWithTx, outboxMessage)
	if err != nil {
		h.logger.Println("failed to publish outbox message", "error", err)
		return err
	}

	err = h.inbox.MarkAsCompleted(ctxWithTx, cmd.ID)
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

func (h *CommandHandler) handleCreateOrder(ctx context.Context, jsonPayload json.RawMessage) (event.Event, error) {
	h.logger.Printf("Handle create order: %+v", jsonPayload)
	var e event.Event

	var payload command.CreateOrderPayload
	err := json.Unmarshal(jsonPayload, &payload)
	if err != nil {
		h.logger.Printf("Error unmarshalling payload: %s", err)
		return e, err
	}

	timeNow := time.Now()
	orderID := uuid.New().String()
	var items []model.OrderItem
	for _, item := range payload.OrderItems {
		items = append(items, model.OrderItem{
			ID:        uuid.New().String(),
			OrderID:   orderID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		})
	}
	order := model.Order{
		ID:              orderID,
		UserID:          payload.UserID,
		PaymentMethodID: payload.PaymentMethodID,
		Phone:           payload.Phone,
		Email:           payload.Email,
		Status:          model.OrderStatusCreated,
		Items:           items,
		CreatedAt:       timeNow,
		UpdatedAt:       timeNow,
	}
	o, e, err := h.orderService.Store(ctx, order)
	if err != nil {
		h.logger.Printf("Error storing order: %s", err)
		return e, err
	}
	h.logger.Printf("Order stored with id: %s", o.ID)

	return e, nil
}

func (h *CommandHandler) handleCompleteOrder(ctx context.Context, jsonPayload json.RawMessage) (event.Event, error) {
	h.logger.Printf("Handle complete order: %+v", jsonPayload)
	var e event.Event

	var payload command.CompleteOrderPayload
	err := json.Unmarshal(jsonPayload, &payload)
	if err != nil {
		h.logger.Printf("Error unmarshalling payload: %s", err)
		return e, err
	}

	e, err = h.orderService.Complete(ctx, payload.OrderID)
	if err != nil {
		h.logger.Printf("Error storing order: %s", err)
		return e, err
	}
	h.logger.Printf("Order completed with id: %s", payload.OrderID)

	return e, nil
}
