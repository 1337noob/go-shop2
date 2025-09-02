package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"shop/order_saga/internal/orchestrator"
	"shop/pkg/broker"
	"shop/pkg/event"
	"shop/pkg/inbox"
	"shop/pkg/outbox"
	"time"
)

type EventHandler struct {
	db           *sql.DB
	orchestrator *orchestrator.Orchestrator
	inbox        inbox.Inbox
	outbox       outbox.Outbox
	logger       *log.Logger
}

func NewEventHandler(db *sql.DB, orchestrator *orchestrator.Orchestrator, inbox inbox.Inbox, outbox outbox.Outbox, logger *log.Logger) *EventHandler {
	return &EventHandler{
		db:           db,
		orchestrator: orchestrator,
		inbox:        inbox,
		outbox:       outbox,
		logger:       logger,
	}
}

func (h *EventHandler) Handle(message broker.Message) error {
	h.logger.Printf("Handling message %s from %s", message.Key, message.Topic)
	h.logger.Println(string(message.Value))

	type eventWithRawPayload struct {
		ID      string          `json:"event_id"`
		Type    event.Type      `json:"event_type"`
		SagaID  string          `json:"saga_id"`
		Payload json.RawMessage `json:"payload"`
	}

	var e eventWithRawPayload
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

	var ev event.Event
	var p any
	switch e.Type {

	case event.OrderCreated:
		h.logger.Printf("Order created event: %+v", e)
		p = event.OrderCreatedPayload{}
		err = json.Unmarshal(e.Payload, &p)
		if err != nil {
			h.logger.Printf("Error unmarshalling payload: %s", err)
			return err
		}

	case event.OrderCreateFailed:
		h.logger.Printf("Order create failed event: %+v", e)
		p = event.OrderCreateFailedPayload{}
		err = json.Unmarshal(e.Payload, &p)
		if err != nil {
			h.logger.Printf("Error unmarshalling payload: %s", err)
			return err
		}

	default:
		return errors.New("invalid event type")

	}

	ev = event.Event{
		ID:      e.ID,
		Type:    e.Type,
		SagaID:  e.SagaID,
		Payload: p,
	}

	err = h.orchestrator.HandleEvent(ctxWithTx, ev)
	if err != nil {
		h.logger.Printf("Error orc handling event: %s", err)
		return err
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
