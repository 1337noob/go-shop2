package service

import (
	"log"
	"shop/product/internal/repository"
)

type CategoryService struct {
	repo   repository.CategoryRepository
	logger *log.Logger
}

func NewCategoryService(repo repository.CategoryRepository, logger *log.Logger) *CategoryService {
	return &CategoryService{repo: repo, logger: logger}
}
