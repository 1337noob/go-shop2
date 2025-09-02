package repository

import (
	"context"
	"database/sql"
	"errors"
	"shop/inventory/internal/model"
	"time"
)

type InventoryRepository interface {
	Create(ctx context.Context, inventory model.Inventory) (model.Inventory, error)
	FindByProductID(ctx context.Context, productID string) (model.Inventory, error)
	GetAvailableQuantity(ctx context.Context, productID string) (int, error)
	Reserve(ctx context.Context, items []model.Item) error
	Release(ctx context.Context, items []model.Item) error
}

type PostgresInventoryRepository struct{}

func NewPostgresInventoryRepository() *PostgresInventoryRepository {
	return &PostgresInventoryRepository{}
}

func (r *PostgresInventoryRepository) Create(ctx context.Context, inventory model.Inventory) (model.Inventory, error) {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return model.Inventory{}, errors.New("transaction not found in context")
	}

	timeNow := time.Now()
	q := `INSERT INTO inventory (product_id, quantity, created_at, updated_at) VALUES ($1, $2, $3, $4)`
	_, err := tx.ExecContext(ctx, q, inventory.ProductID, inventory.Quantity, timeNow, timeNow)
	if err != nil {
		return model.Inventory{}, err
	}

	return inventory, nil
}

func (r *PostgresInventoryRepository) FindByProductID(ctx context.Context, productID string) (model.Inventory, error) {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return model.Inventory{}, errors.New("transaction not found in context")
	}
	q := `SELECT product_id, quantity, created_at, updated_at FROM inventory WHERE product_id = $1`
	var inventory model.Inventory
	err := tx.QueryRowContext(ctx, q, productID).Scan(
		&inventory.ProductID, &inventory.Quantity, &inventory.CreatedAt, &inventory.UpdatedAt,
	)
	if err != nil {
		return model.Inventory{}, err
	}

	return inventory, nil
}

func (r *PostgresInventoryRepository) GetAvailableQuantity(ctx context.Context, productID string) (int, error) {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return 0, errors.New("transaction not found in context")
	}

	q := `SELECT quantity FROM inventory WHERE product_id = $1`
	var quantity int
	err := tx.QueryRowContext(ctx, q, productID).Scan(&quantity)
	if err != nil {
		return 0, err
	}

	return quantity, nil
}

func (r *PostgresInventoryRepository) Reserve(ctx context.Context, items []model.Item) error {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return errors.New("transaction not found in context")
	}

	for _, item := range items {
		q := `UPDATE inventory SET quantity = quantity - $1, updated_at = $2 WHERE product_id = $3 AND quantity >= $1`
		result, err := tx.ExecContext(ctx, q, item.Quantity, time.Now(), item.ProductID)
		if err != nil {
			return err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			return errors.New("inventory rows affected is zero")
		}
	}

	return nil
}

func (r *PostgresInventoryRepository) Release(ctx context.Context, items []model.Item) error {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return errors.New("transaction not found in context")
	}

	for _, item := range items {
		q := `UPDATE inventory SET quantity = quantity + $1, updated_at = $2 WHERE product_id = $3`
		result, err := tx.ExecContext(ctx, q, item.Quantity, time.Now(), item.ProductID)
		if err != nil {
			return err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			return errors.New("inventory rows affected is zero")
		}
	}

	return nil
}
