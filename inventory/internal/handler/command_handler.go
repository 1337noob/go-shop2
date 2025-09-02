package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"log"
	"shop/inventory/internal/model"
	"shop/inventory/internal/service"
	"shop/pkg/broker"
	"shop/pkg/command"
	"shop/pkg/event"
	"shop/pkg/inbox"
	"shop/pkg/outbox"
	"time"
)

type CommandHandler struct {
	db               *sql.DB
	inventoryService *service.InventoryService
	inbox            inbox.Inbox
	outbox           outbox.Outbox
	logger           *log.Logger
}

func NewCommandHandler(db *sql.DB, inventoryService *service.InventoryService, inbox inbox.Inbox, outbox outbox.Outbox, logger *log.Logger) *CommandHandler {
	return &CommandHandler{
		db:               db,
		inventoryService: inventoryService,
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

	var e event.Event

	switch cmd.Type {
	case command.CreateInventory:
		h.logger.Printf("Create inventory command: %+v", cmd)
		e, err = h.handleCreateInventory(ctxWithTx, cmd.Payload)
		if err != nil {
			return err
		}
	case command.ReserveInventory:
		h.logger.Printf("Reserve products command: %+v", cmd)
		e, err = h.handleReserve(ctxWithTx, cmd)
		if err != nil {
			return err
		}
	case command.ReleaseInventory:
		h.logger.Printf("Release products command: %+v", cmd)
		e, err = h.handleRelease(ctxWithTx, cmd)
		if err != nil {
			return err
		}
	default:
		return errors.New("invalid command")
	}

	e.SagaID = cmd.SagaID

	outboxMessage := outbox.Message{
		ID:        uuid.New().String(),
		Topic:     "inventory-events",
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

func (h *CommandHandler) handleCreateInventory(ctx context.Context, jsonPayload json.RawMessage) (event.Event, error) {
	h.logger.Printf("Handle create inventory: %+v", jsonPayload)
	var e event.Event

	var payload command.CreateInventoryPayload
	err := json.Unmarshal(jsonPayload, &payload)
	if err != nil {
		h.logger.Printf("Error unmarshalling payload: %s", err)
		return e, err
	}

	inventory := model.Inventory{
		ProductID: payload.ProductID,
		Quantity:  payload.Quantity,
		//ReservedQuantity: payload.ReservedQuantity,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	inv, e, err := h.inventoryService.Store(ctx, inventory)
	if err != nil {
		h.logger.Printf("Error storing inventory: %s", err)
		return e, err
	}
	h.logger.Printf("Inventory stored with id: %s", inv.ProductID)

	return e, nil
}

func (h *CommandHandler) handleReserve(ctx context.Context, cmd command.Command) (event.Event, error) {
	h.logger.Printf("Handle reserve products: %+v", cmd)
	var e event.Event

	var payload command.ReserveProductsPayload
	err := json.Unmarshal(cmd.Payload, &payload)
	if err != nil {
		h.logger.Printf("Error unmarshalling payload: %s", err)
		return e, err
	}
	var items []model.Item
	for _, item := range payload.Items {
		items = append(items, model.Item{
			ProductID: item.ProductId,
			Quantity:  item.Quantity,
		})
	}

	e, err = h.inventoryService.Reserve(ctx, cmd.SagaID, items)
	if err != nil {
		h.logger.Printf("Error reserve inventory: %s", err)
		return e, err
	}

	return e, nil
}

func (h *CommandHandler) handleRelease(ctx context.Context, cmd command.Command) (event.Event, error) {
	h.logger.Printf("Handle release products: %+v", cmd)
	var e event.Event

	var payload command.ReleaseProductsPayload
	err := json.Unmarshal(cmd.Payload, &payload)
	if err != nil {
		h.logger.Printf("Error unmarshalling payload: %s", err)
		return e, err
	}

	var items []model.Item
	for _, item := range payload.Items {
		items = append(items, model.Item{
			ProductID: item.ProductId,
			Quantity:  item.Quantity,
		})
	}

	e, err = h.inventoryService.Release(ctx, cmd.SagaID, items)
	if err != nil {
		h.logger.Printf("Error release inventory: %s", err)
		return e, err
	}

	return e, nil
}
