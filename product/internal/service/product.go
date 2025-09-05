package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"shop/pkg/event"
	"shop/pkg/types"
	"shop/product/internal/repository"

	"github.com/google/uuid"
)

type ProductService struct {
	repo   repository.ProductRepository
	logger *log.Logger
}

func NewProductService(repo repository.ProductRepository, logger *log.Logger) *ProductService {
	return &ProductService{repo: repo, logger: logger}
}

func (s *ProductService) ValidateProductsByIds(ctx context.Context, items []types.Item) (event.Event, error) {
	var e event.Event
	var validatedItems []types.Item

	eventId := uuid.New().String()

	for _, item := range items {
		product, err := s.repo.FindById(ctx, item.ProductID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				s.logger.Println("failed to find product", "error", err)
				jsonPayload, err := json.Marshal(event.ProductsValidationFailedPayload{
					OrderItems: items,
					Error:      err.Error(),
				})
				if err != nil {
					s.logger.Println("failed to marshal payload", "error", err)
					return e, err
				}
				e = event.Event{
					ID:      eventId,
					Type:    event.ProductsValidationFailed,
					Payload: jsonPayload,
				}
			}

			s.logger.Println("failed to find product", "error", err)
			return e, err
		}

		validatedItems = append(validatedItems, types.Item{
			ProductID: product.ID,
			Name:      product.Name,
			Price:     product.Price,
		})
	}

	p := event.ProductsValidatedPayload{
		OrderItems: validatedItems,
	}
	jsonPayload, err := json.Marshal(p)
	if err != nil {
		s.logger.Println("failed to marshal payload", "error", err)
		return e, err
	}
	e = event.Event{
		ID:      eventId,
		Type:    event.ProductsValidated,
		Payload: jsonPayload,
	}

	return e, nil
}
