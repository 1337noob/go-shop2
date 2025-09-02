package repository

import (
	"context"
	"database/sql"
	"errors"
	"shop/product/internal/model"
)

type ProductRepository interface {
	Create(ctx context.Context, product model.Product) (model.Product, error)
	FindById(ctx context.Context, id string) (model.Product, error)
}

type PostgresProductRepository struct{}

func NewPostgresProductRepository() *PostgresProductRepository {
	return &PostgresProductRepository{}
}

func (p *PostgresProductRepository) Create(ctx context.Context, product model.Product) (model.Product, error) {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return model.Product{}, errors.New("transaction not found in context")
	}

	_, err := tx.ExecContext(ctx, "INSERT INTO products (id, name, price, category_id) VALUES ($1, $2, $3, $4)", product.ID, product.Name, product.Price, product.CategoryID)
	if err != nil {
		return model.Product{}, err
	}

	return product, nil
}

func (p *PostgresProductRepository) FindById(ctx context.Context, id string) (model.Product, error) {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return model.Product{}, errors.New("transaction not found in context")
	}

	var product model.Product
	err := tx.QueryRowContext(ctx, "SELECT id, name, price, category_id FROM products WHERE id=$1", id).Scan(&product.ID, &product.Name, &product.Price, &product.CategoryID)
	if err != nil {
		return model.Product{}, err
	}

	return product, nil
}
