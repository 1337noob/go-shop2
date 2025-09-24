package service

import (
	"context"
	"encoding/json"
	"log"
	"shop/inventory/internal/model"
	"shop/inventory/internal/repository"
	"shop/pkg/event"
	"shop/pkg/types"

	"github.com/google/uuid"
)

type InventoryService struct {
	repo   repository.InventoryRepository
	logger *log.Logger
}

func NewInventoryService(repo repository.InventoryRepository, logger *log.Logger) *InventoryService {
	return &InventoryService{repo: repo, logger: logger}
}

func (s *InventoryService) Reserve(ctx context.Context, sagaID string, items []model.Item, orderID string) (event.Event, error) {
	var e event.Event

	err := s.repo.Reserve(ctx, items)
	var eventItems []types.Item
	for _, item := range items {
		eventItems = append(eventItems, types.Item{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	eventID := uuid.New().String()

	if err != nil {
		s.logger.Println("failed to reserve inventory", "error", err)
		// TODO event InventoryReserveFailed

		payload := event.InventoryReserveFailedPayload{
			OrderID:    orderID,
			OrderItems: eventItems,
			Error:      err.Error(),
		}
		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			s.logger.Println("failed to marshal payload", "error", err)
			return e, err
		}
		e = event.Event{
			ID:      eventID,
			Type:    event.InventoryReserveFailed,
			SagaID:  sagaID,
			Payload: jsonPayload,
		}
		return e, err
	}

	payload := event.InventoryReservedPayload{
		OrderID:    orderID,
		OrderItems: eventItems,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		s.logger.Println("failed to marshal payload", "error", err)
		return e, err
	}
	e = event.Event{
		ID:      eventID,
		Type:    event.InventoryReserved,
		SagaID:  sagaID,
		Payload: jsonPayload,
	}

	return e, nil
}

func (s *InventoryService) Release(ctx context.Context, sagaID string, items []model.Item, orderID string) (event.Event, error) {
	var e event.Event

	err := s.repo.Release(ctx, items)

	eventID := uuid.New().String()

	var eventItems []types.Item
	for _, item := range items {
		eventItems = append(eventItems, types.Item{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	if err != nil {
		s.logger.Println("failed to release inventory", "error", err)
		// TODO event InventoryReleaseFailed
		payload := event.InventoryReleaseFailedPayload{
			OrderID:    orderID,
			OrderItems: eventItems,
			Error:      err.Error(),
		}
		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			s.logger.Println("failed to marshal payload", "error", err)
			return e, err
		}
		e = event.Event{
			ID:      eventID,
			Type:    event.InventoryReleaseFailed,
			SagaID:  sagaID,
			Payload: jsonPayload,
		}
		return e, err
	}

	payload := event.InventoryReleasedPayload{
		OrderID:    orderID,
		OrderItems: eventItems,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		s.logger.Println("failed to marshal payload", "error", err)
		return e, err
	}
	e = event.Event{
		ID:      eventID,
		Type:    event.InventoryReleased,
		SagaID:  sagaID,
		Payload: jsonPayload,
	}

	return e, nil
}
