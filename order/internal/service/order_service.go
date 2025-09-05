package service

import (
	"context"
	"encoding/json"
	"log"
	"shop/order/internal/model"
	"shop/order/internal/repository"
	"shop/pkg/event"
	"shop/pkg/types"

	"github.com/google/uuid"
)

type OrderService struct {
	repo   repository.OrderRepository
	logger *log.Logger
}

func NewOrderService(repo repository.OrderRepository, logger *log.Logger) *OrderService {
	return &OrderService{repo: repo, logger: logger}
}

func (s *OrderService) Store(ctx context.Context, order model.Order) (model.Order, event.Event, error) {
	var e event.Event

	o, err := s.repo.Create(ctx, order)
	if err != nil {
		s.logger.Println("failed to create order", "error", err)
		return model.Order{}, e, err
	}

	var items []types.Item
	for _, item := range o.Items {
		items = append(items, types.Item{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}
	payload := event.OrderCreatedPayload{
		OrderID:         o.ID,
		UserID:          o.UserID,
		PaymentMethodID: o.PaymentMethodID,
		Phone:           o.Phone,
		Email:           o.Email,
		Status:          string(o.Status),
		OrderItems:      items,
		CreatedAt:       o.CreatedAt,
		UpdatedAt:       o.UpdatedAt,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		s.logger.Println("failed to marshal payload", "error", err)
		return model.Order{}, e, err
	}
	e = event.Event{
		ID:      uuid.New().String(),
		Type:    event.OrderCreated,
		Payload: jsonPayload,
	}

	return o, e, nil
}

func (s *OrderService) Complete(ctx context.Context, orderID string) (event.Event, error) {
	var e event.Event

	newStatus := model.OrderStatusCompleted
	err := s.repo.UpdateStatus(ctx, orderID, newStatus)
	if err != nil {
		s.logger.Println("failed to complete order", "error", err)
		return e, err
	}

	payload := event.OrderCompletedPayload{
		OrderID: orderID,
		Status:  string(newStatus),
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		s.logger.Println("failed to marshal payload", "error", err)
		return e, err
	}
	e = event.Event{
		ID:      uuid.New().String(),
		Type:    event.OrderCompleted,
		Payload: jsonPayload,
	}

	return e, nil
}

func (s *OrderService) Cancel(ctx context.Context, orderID string) (event.Event, error) {
	var e event.Event

	newStatus := model.OrderStatusCancelled
	err := s.repo.UpdateStatus(ctx, orderID, newStatus)
	if err != nil {
		s.logger.Println("failed to complete order", "error", err)
		return e, err
	}

	payload := event.OrderCancelledPayload{
		OrderID: orderID,
		Status:  string(newStatus),
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		s.logger.Println("failed to marshal payload", "error", err)
		return e, err
	}
	e = event.Event{
		ID:      uuid.New().String(),
		Type:    event.OrderCancelled,
		Payload: jsonPayload,
	}

	return e, nil
}
