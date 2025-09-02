package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"log"
	"shop/pkg/broker"
	"shop/pkg/command"
	"shop/pkg/event"
	"shop/pkg/inbox"
	"shop/pkg/outbox"
	"shop/product/internal/model"
	"shop/product/internal/service"
	"time"
)

type CommandHandler struct {
	db              *sql.DB
	categoryService *service.CategoryService
	productService  *service.ProductService
	inbox           inbox.Inbox
	outbox          outbox.Outbox
	logger          *log.Logger
}

func NewCommandHandler(db *sql.DB, categoryService *service.CategoryService, productService *service.ProductService, inbox inbox.Inbox, outbox outbox.Outbox, logger *log.Logger) *CommandHandler {
	return &CommandHandler{
		db:              db,
		categoryService: categoryService,
		productService:  productService,
		inbox:           inbox,
		outbox:          outbox,
		logger:          logger,
	}
}

func (h *CommandHandler) Handle(message broker.Message) error {
	h.logger.Printf("Handling message %s from %s", message.Key, message.Topic)

	var cmd command.Command
	err := json.Unmarshal(message.Value, &cmd)
	if err != nil {
		h.logger.Printf("Error unmarshalling command: %s", err)
		return err
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
	case command.CreateCategory:
		h.logger.Printf("Create category command: %+v", cmd)
		e, err = h.handleCreateCategory(ctxWithTx, cmd.Payload)
		if err != nil {
			return err
		}
		h.logger.Printf("Create category complete")
	case command.CreateProduct:
		h.logger.Printf("Handling create product: %+v", cmd)
		e, err = h.handleCreateProduct(ctxWithTx, cmd.Payload)
		if err != nil {
			return err
		}
		h.logger.Printf("Create product complete")
	case command.ValidateProducts:
		h.logger.Printf("Handling validate products: %+v", cmd)
		e, err = h.handleValidateProducts(ctxWithTx, cmd.Payload)
		if err != nil {
			return err
		}
		h.logger.Printf("Validate products complete")
	default:
		return errors.New("invalid command")
	}

	e.SagaID = cmd.SagaID

	outboxMessage := outbox.Message{
		ID:        uuid.New().String(),
		Topic:     "product-events",
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

func (h *CommandHandler) handleCreateCategory(ctx context.Context, jsonPayload json.RawMessage) (event.Event, error) {
	h.logger.Printf("Handle create category: %+v", jsonPayload)
	var e event.Event

	var payload command.CreateCategoryPayload
	err := json.Unmarshal(jsonPayload, &payload)
	if err != nil {
		h.logger.Printf("Error unmarshalling payload: %s", err)
		return e, err
	}

	category := model.Category{
		ID:   uuid.New().String(),
		Name: payload.Name,
	}
	cat, e, err := h.categoryService.Store(ctx, category)
	if err != nil {
		h.logger.Printf("Failed to store category: %s", err)
		return e, err
	}
	h.logger.Printf("Created category with id: %s", cat.ID)

	return e, nil
}

func (h *CommandHandler) handleCreateProduct(ctx context.Context, jsonPayload json.RawMessage) (event.Event, error) {
	h.logger.Printf("Handle create product: %+v", jsonPayload)
	var e event.Event

	var payload command.CreateProductPayload
	err := json.Unmarshal(jsonPayload, &payload)
	if err != nil {
		h.logger.Printf("Error unmarshalling payload: %s", err)
		return e, err
	}

	product := model.Product{
		ID:         uuid.New().String(),
		Name:       payload.Name,
		Price:      payload.Price,
		CategoryID: payload.CategoryID,
	}
	prod, e, err := h.productService.Store(ctx, product)
	if err != nil {
		h.logger.Printf("Failed to store product: %s", err)
		return e, err
	}
	h.logger.Printf("Created product with id: %s", prod.ID)

	return e, nil
}

func (h *CommandHandler) handleValidateProducts(ctx context.Context, jsonPayload json.RawMessage) (event.Event, error) {
	h.logger.Printf("Handle create product: %+v", jsonPayload)
	var e event.Event

	var payload command.ValidateProductsPayload
	err := json.Unmarshal(jsonPayload, &payload)
	if err != nil {
		h.logger.Printf("Error unmarshalling payload: %s", err)
		return e, err
	}

	var productIDs []string
	for _, pr := range payload.OrderItems {
		productIDs = append(productIDs, pr.ID)
	}

	e, err = h.productService.ValidateProductsByIds(ctx, productIDs)
	if err != nil {
		h.logger.Printf("Failed to validate products: %s", err)
		return e, err
	}

	return e, nil
}
