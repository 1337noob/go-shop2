package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"shop/gateway/internal/middleware"
	"shop/pkg/command"
	"shop/pkg/outbox"
	"shop/pkg/types"
	"time"

	"github.com/google/uuid"
)

type OrderHandler struct {
	db     *sql.DB
	outbox outbox.Outbox
	logger *log.Logger
}

func NewOrderHandler(db *sql.DB, outbox outbox.Outbox, logger *log.Logger) *OrderHandler {
	return &OrderHandler{
		db:     db,
		outbox: outbox,
		logger: logger,
	}
}

type CreateOrderRequest struct {
	PaymentMethodID string       `json:"payment_method_id"`
	OrderItems      []types.Item `json:"order_items"`
}

func (o *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	o.logger.Println("CreateOrder handler start")

	session := middleware.GetSessionFromContext(r.Context())
	if session == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateOrderRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	payload := command.SagaCreateOrderPayload{
		UserID:          session.UserID,
		PaymentMethodID: req.PaymentMethodID,
		OrderItems:      req.OrderItems,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	cmd := command.Command{
		ID:      uuid.New().String(),
		Type:    command.SagaCreateOrder,
		Payload: jsonPayload,
	}

	outboxMessage := outbox.Message{
		ID:        uuid.New().String(),
		Topic:     "order-saga-commands",
		Payload:   cmd,
		Status:    outbox.StatusInit,
		CreatedAt: time.Now(),
	}

	tx, err := o.db.Begin()
	if err != nil {
		o.logger.Println("Failed to begin transaction", "error", err)
		http.Error(w, "Failed to begin transaction", http.StatusInternalServerError)
	}
	defer tx.Rollback()

	ctxWithTx := context.WithValue(context.Background(), "tx", tx)

	err = o.outbox.Publish(ctxWithTx, outboxMessage)
	if err != nil {
		o.logger.Println("failed to publish outbox message", "error", err)
		http.Error(w, "failed to publish outbox message", http.StatusInternalServerError)
	}

	err = tx.Commit()
	if err != nil {
		o.logger.Println("failed to commit transaction", "error", err)
	}

	o.logger.Println("CreateOrder handler finished")
}
