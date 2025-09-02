package service

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"log"
	"shop/pkg/event"
	"shop/product/internal/model"
	"shop/product/internal/repository"
)

type ProductService struct {
	repo   repository.ProductRepository
	logger *log.Logger
}

func NewProductService(repo repository.ProductRepository, logger *log.Logger) *ProductService {
	return &ProductService{repo: repo, logger: logger}
}

func (s *ProductService) Store(ctx context.Context, product model.Product) (model.Product, event.Event, error) {
	var e event.Event

	prod, err := s.repo.Create(ctx, product)
	if err != nil {
		s.logger.Println("failed to create product", "error", err)
		return model.Product{}, e, err
	}

	p := event.ProductCreatedPayload{
		ID:         prod.ID,
		Name:       prod.Name,
		Price:      prod.Price,
		CategoryID: prod.CategoryID,
	}
	e = event.Event{
		ID:      uuid.New().String(),
		Type:    event.TypeProductCreated,
		Payload: p,
	}

	return prod, e, nil
}

func (s *ProductService) ValidateProductsByIds(ctx context.Context, ids []string) (event.Event, error) {
	var e event.Event
	var validatedProducts []event.Product

	eventId := uuid.New().String()

	for _, id := range ids {
		prod, err := s.repo.FindById(ctx, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				s.logger.Println("failed to find product", "error", err)
				e = event.Event{
					ID:   eventId,
					Type: event.TypeProductsValidationFailed,
					Payload: event.ProductsValidationFailedPayload{
						Error: err.Error(),
					},
				}
			}

			s.logger.Println("failed to find product", "error", err)
			return e, err
		}

		validatedProducts = append(validatedProducts, event.Product{
			ID:    prod.ID,
			Name:  prod.Name,
			Price: prod.Price,
		})
	}

	p := event.ProductsValidatedPayload{
		Products: validatedProducts,
	}
	e = event.Event{
		ID:      eventId,
		Type:    event.TypeProductsValidated,
		Payload: p,
	}

	return e, nil
}
