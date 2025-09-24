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
	Get(ctx context.Context, page int, limit int) ([]model.Category, error)
}

type PostgresCategoryRepository struct {
	db *sql.DB
}

func NewPostgresCategoryRepository(db *sql.DB) *PostgresCategoryRepository {
	return &PostgresCategoryRepository{
		db: db,
	}
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

func (r *PostgresCategoryRepository) Get(ctx context.Context, page int, limit int) ([]model.Category, error) {
	//tx, ok := ctx.Value("tx").(*sql.Tx)
	//if !ok {
	//	return []model.Category{}, errors.New("transaction not found in context")
	//}
	//
	var categories []model.Category
	offset := (page - 1) * limit
	query := `SELECT id, name, created_at  FROM categories ORDER BY created_at DESC OFFSET $1 LIMIT $2`
	rows, err := r.db.QueryContext(ctx, query, offset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var c model.Category
		err = rows.Scan(&c.ID, &c.Name, &c.CreatedAt)
		if err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}

	return categories, nil
}
