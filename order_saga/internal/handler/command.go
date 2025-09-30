package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"shop/order_saga/internal/service"
	"shop/pkg/broker"
	"shop/pkg/command"
	"shop/pkg/inbox"
	"shop/pkg/outbox"
	"time"
)

type CommandHandler struct {
	db               *sql.DB
	orderSagaService *service.OrderSagaService
	inbox            inbox.Inbox
	outbox           outbox.Outbox
	logger           *log.Logger
}

func NewCommandHandler(db *sql.DB, orderSagaService *service.OrderSagaService, inbox inbox.Inbox, outbox outbox.Outbox, logger *log.Logger) *CommandHandler {
	return &CommandHandler{
		db:               db,
		orderSagaService: orderSagaService,
		inbox:            inbox,
		outbox:           outbox,
		logger:           logger,
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

	switch cmd.Type {
	case command.SagaCreateOrder:
		h.logger.Printf("Create order command: %+v", cmd)
		err = h.handleSagaCreateOrder(ctxWithTx, cmd.Payload)
		if err != nil {
			return err
		}
	default:
		return errors.New("invalid command")
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

func (h *CommandHandler) handleSagaCreateOrder(ctx context.Context, jsonPayload json.RawMessage) error {
	h.logger.Printf("Handle create order: %+v", jsonPayload)

	var payload command.SagaCreateOrderPayload
	err := json.Unmarshal(jsonPayload, &payload)
	if err != nil {
		h.logger.Printf("Error unmarshalling payload: %s", err)
		return err
	}

	err = h.orderSagaService.Create(ctx, payload.UserID, payload.OrderItems, payload.PaymentMethodID)
	if err != nil {
		h.logger.Printf("Error storing order: %s", err)
		return err
	}
	h.logger.Println("Create order saga created successfully")

	return nil
}
