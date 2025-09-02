package service

import (
	"context"
	"github.com/google/uuid"
	"log"
	"shop/pkg/event"
	"shop/product/internal/model"
	"shop/product/internal/repository"
)

type CategoryService struct {
	repo   repository.CategoryRepository
	logger *log.Logger
}

func NewCategoryService(repo repository.CategoryRepository, logger *log.Logger) *CategoryService {
	return &CategoryService{repo: repo, logger: logger}
}

func (s *CategoryService) Store(ctx context.Context, category model.Category) (model.Category, event.Event, error) {
	var e event.Event

	cat, err := s.repo.Create(ctx, category)
	if err != nil {
		s.logger.Println("failed to create category", "error", err)
		return model.Category{}, e, err
	}

	p := event.CategoryCreatedPayload{
		ID:   cat.ID,
		Name: cat.Name,
	}
	e = event.Event{
		ID:      uuid.New().String(),
		Type:    event.TypeCategoryCreated,
		Payload: p,
	}

	return cat, e, nil
}
