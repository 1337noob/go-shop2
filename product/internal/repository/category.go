package repository

import (
	"context"
	"database/sql"
	"errors"
	"shop/product/internal/model"
)

type CategoryRepository interface {
	Create(ctx context.Context, category model.Category) (model.Category, error)
	FindByID(ctx context.Context, id string) (model.Category, error)
}

type PostgresCategoryRepository struct{}

func NewPostgresCategoryRepository() *PostgresCategoryRepository {
	return &PostgresCategoryRepository{}
}

func (r *PostgresCategoryRepository) Create(ctx context.Context, category model.Category) (model.Category, error) {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return model.Category{}, errors.New("transaction not found in context")
	}

	_, err := tx.ExecContext(ctx, "INSERT INTO categories (id, name) VALUES ($1, $2)", category.ID, category.Name)
	if err != nil {
		return model.Category{}, err
	}

	return category, nil
}

func (r *PostgresCategoryRepository) FindByID(ctx context.Context, id string) (model.Category, error) {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return model.Category{}, errors.New("transaction not found in context")
	}

	var category model.Category
	err := tx.QueryRowContext(ctx, "SELECT id, name  FROM categories WHERE id = $1", id).Scan(
		&category.ID, &category.Name,
	)
	if err != nil {
		return model.Category{}, err
	}

	return category, nil
}
