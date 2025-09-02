package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"log"
	"shop/payment/internal/model"
	"shop/payment/internal/service"
	"shop/pkg/broker"
	"shop/pkg/command"
	"shop/pkg/event"
	"shop/pkg/inbox"
	"shop/pkg/outbox"
	"time"
)

type CommandHandler struct {
	db             *sql.DB
	methodService  *service.MethodService
	paymentService *service.PaymentService
	inbox          inbox.Inbox
	outbox         outbox.Outbox
	logger         *log.Logger
}

func NewCommandHandler(db *sql.DB, methodService *service.MethodService, paymentService *service.PaymentService, inbox inbox.Inbox, outbox outbox.Outbox, logger *log.Logger) *CommandHandler {
	return &CommandHandler{
		db:             db,
		methodService:  methodService,
		paymentService: paymentService,
		inbox:          inbox,
		outbox:         outbox,
		logger:         logger,
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
	case command.CreatePaymentMethod:
		h.logger.Printf("Create payment method command: %+v", cmd)
		e, err = h.handleCreatePaymentMethod(ctxWithTx, cmd.Payload)
		if err != nil {
			return err
		}
	case command.ProcessPayment:
		h.logger.Printf("Process payment command: %+v", cmd)
		e, err = h.handleProcessPayment(ctxWithTx, cmd.Payload)
		if err != nil {
			return err
		}
	default:
		return errors.New("invalid command")
	}

	e.SagaID = cmd.SagaID

	outboxMessage := outbox.Message{
		ID:        uuid.New().String(),
		Topic:     "payment-events",
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

func (h *CommandHandler) handleCreatePaymentMethod(ctx context.Context, jsonPayload json.RawMessage) (event.Event, error) {
	h.logger.Printf("Handle create payment method: %+v", jsonPayload)
	var e event.Event

	var payload command.CreatePaymentMethodPayload
	err := json.Unmarshal(jsonPayload, &payload)
	if err != nil {
		h.logger.Printf("Error unmarshalling payload: %s", err)
		return e, err
	}

	method := model.Method{
		ID:          uuid.New().String(),
		UserID:      payload.UserID,
		Gateway:     payload.Gateway,
		PaymentType: payload.PaymentType,
		Token:       payload.Token,
	}
	m, e, err := h.methodService.Store(ctx, method)
	if err != nil {
		h.logger.Printf("Error storing payment method: %s", err)
		return e, err
	}
	h.logger.Printf("Payment method stored with id: %s", m.ID)

	return e, nil
}

func (h *CommandHandler) handleProcessPayment(ctx context.Context, jsonPayload json.RawMessage) (event.Event, error) {
	h.logger.Printf("Handle create payment method: %+v", jsonPayload)
	var e event.Event

	var payload command.ProcessPaymentPayload
	err := json.Unmarshal(jsonPayload, &payload)
	if err != nil {
		h.logger.Printf("Error unmarshalling payload: %s", err)
		return e, err
	}

	payment := model.Payment{
		ID:         uuid.New().String(),
		OrderID:    payload.OrderID,
		UserID:     payload.UserID,
		Amount:     payload.Amount,
		ExternalID: "",
		Status:     model.PaymentStatusPending,
		MethodID:   payload.MethodID,
	}
	pay, e, err := h.paymentService.Process(ctx, payment)
	if err != nil {
		h.logger.Printf("Error storing payment method: %s", err)
		return e, err
	}
	h.logger.Printf("Payment processed with id: %s", pay.ID)

	return e, nil
}
