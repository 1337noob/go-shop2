package service

import (
	"context"
	"github.com/google/uuid"
	"log"
	"shop/inventory/internal/model"
	"shop/inventory/internal/repository"
	"shop/pkg/event"
)

type InventoryService struct {
	repo   repository.InventoryRepository
	logger *log.Logger
}

func NewInventoryService(repo repository.InventoryRepository, logger *log.Logger) *InventoryService {
	return &InventoryService{repo: repo, logger: logger}
}

func (s *InventoryService) Store(ctx context.Context, inventory model.Inventory) (model.Inventory, event.Event, error) {
	var e event.Event

	inv, err := s.repo.Create(ctx, inventory)
	if err != nil {
		s.logger.Println("failed to create inventory", "error", err)
		return model.Inventory{}, e, err
	}

	eventID := uuid.New().String()
	payload := event.InventoryCreatedPayload{
		ProductID: inv.ProductID,
		Quantity:  inv.Quantity,
		//ReservedQuantity: inv.ReservedQuantity,
		CreatedAt: inv.CreatedAt,
		UpdatedAt: inv.UpdatedAt,
	}
	e = event.Event{
		ID:      eventID,
		Type:    event.TypeInventoryCreated,
		Payload: payload,
	}

	return inv, e, nil
}

func (s *InventoryService) Reserve(ctx context.Context, sagaID string, items []model.Item) (event.Event, error) {
	var e event.Event

	err := s.repo.Reserve(ctx, items)
	if err != nil {
		s.logger.Println("failed to reserve inventory", "error", err)
		// TODO event InventoryReserveFailed
		return e, err
	}

	eventID := uuid.New().String()

	var eventItems []event.InventoryItem
	for _, item := range items {
		eventItems = append(eventItems, event.InventoryItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}
	payload := event.InventoryReservedPayload{
		Items: eventItems,
	}
	e = event.Event{
		ID:      eventID,
		Type:    event.InventoryReserved,
		SagaID:  sagaID,
		Payload: payload,
	}

	return e, nil
}

func (s *InventoryService) Release(ctx context.Context, sagaID string, items []model.Item) (event.Event, error) {
	var e event.Event

	err := s.repo.Release(ctx, items)
	if err != nil {
		s.logger.Println("failed to release inventory", "error", err)
		// TODO event InventoryReleaseFailed
		return e, err
	}

	eventID := uuid.New().String()

	var eventItems []event.InventoryItem
	for _, item := range items {
		eventItems = append(eventItems, event.InventoryItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}
	payload := event.InventoryReleasedPayload{
		Items: eventItems,
	}
	e = event.Event{
		ID:      eventID,
		Type:    event.InventoryReleased,
		SagaID:  sagaID,
		Payload: payload,
	}

	return e, nil
}
