package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"shop/order_history/internal/model"
)

type OrderRepository interface {
	Create(ctx context.Context, order *model.Order) error
	FindByID(ctx context.Context, id string) (*model.Order, error)
	Update(ctx context.Context, order *model.Order) error
	GetByUserID(ctx context.Context, userID string, page int, limit int) ([]*model.Order, error)
}

type PostgresOrderRepository struct {
	db *sql.DB
}

func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository {
	return &PostgresOrderRepository{
		db: db,
	}
}

func (o *PostgresOrderRepository) Create(ctx context.Context, order *model.Order) error {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return errors.New("transaction not found in context")
	}

	itemsJson, err := json.Marshal(order.OrderItems)
	if err != nil {
		return err
	}

	query := `INSERT INTO order_history (id, user_id, order_items, payment_id, payment_method_id, payment_type, payment_gateway, payment_sum, payment_external_id, payment_status, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`
	_, err = tx.ExecContext(
		ctx,
		query,
		order.ID,
		order.UserID,
		itemsJson,
		order.PaymentID,
		order.PaymentMethodID,
		order.PaymentType,
		order.PaymentGateway,
		order.PaymentSum,
		order.PaymentExternalID,
		order.PaymentStatus,
		order.Status,
		order.CreatedAt,
		order.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (o *PostgresOrderRepository) FindByID(ctx context.Context, id string) (*model.Order, error) {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return nil, errors.New("transaction not found in context")
	}

	var order model.Order
	var itemsJson []byte

	query := `SELECT id, user_id, order_items, payment_id, payment_method_id, payment_type, payment_gateway, payment_sum, payment_external_id, payment_status, status, created_at, updated_at FROM order_history WHERE id = $1`
	err := tx.QueryRowContext(ctx, query, id).Scan(
		&order.ID,
		&order.UserID,
		&itemsJson,
		&order.PaymentID,
		&order.PaymentMethodID,
		&order.PaymentType,
		&order.PaymentGateway,
		&order.PaymentSum,
		&order.PaymentExternalID,
		&order.PaymentStatus,
		&order.Status,
		&order.CreatedAt,
		&order.UpdatedAt,
	)

	err = json.Unmarshal(itemsJson, &order.OrderItems)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (o *PostgresOrderRepository) Update(ctx context.Context, order *model.Order) error {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return errors.New("transaction not found in context")
	}

	itemsJson, err := json.Marshal(order.OrderItems)
	if err != nil {
		return err
	}

	query := `UPDATE order_history SET user_id = $1, order_items = $2, payment_id = $3, payment_method_id = $4, payment_type = $5, payment_gateway = $6, payment_sum = $7, payment_external_id = $8, payment_status = $9, status = $10, created_at = $11, updated_at = $12 WHERE id = $13`
	_, err = tx.ExecContext(
		ctx,
		query,
		order.UserID,
		itemsJson,
		order.PaymentID,
		order.PaymentMethodID,
		order.PaymentType,
		order.PaymentGateway,
		order.PaymentSum,
		order.PaymentExternalID,
		order.PaymentStatus,
		order.Status,
		order.CreatedAt,
		order.UpdatedAt,
		order.ID,
	)

	if err != nil {
		return err
	}

	return nil
}

func (o *PostgresOrderRepository) GetByUserID(ctx context.Context, userID string, page int, limit int) ([]*model.Order, error) {

	offset := (page - 1) * limit

	query := `SELECT id, user_id, order_items, payment_id, payment_method_id, payment_type, payment_gateway, payment_sum, payment_external_id, payment_status, status, created_at, updated_at FROM order_history WHERE user_id = $1 ORDER BY created_at DESC OFFSET $2 LIMIT $3`
	rows, err := o.db.QueryContext(ctx, query, userID, offset, limit)
	if err != nil {
		return nil, err
	}

	var orders []*model.Order

	for rows.Next() {
		var order model.Order
		var itemsJSON []byte
		err = rows.Scan(
			&order.ID,
			&order.UserID,
			&itemsJSON,
			&order.PaymentID,
			&order.PaymentMethodID,
			&order.PaymentType,
			&order.PaymentGateway,
			&order.PaymentSum,
			&order.PaymentExternalID,
			&order.PaymentStatus,
			&order.Status,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(itemsJSON, &order.OrderItems)
		if err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}
	err = rows.Close()
	if err != nil {
		return nil, err
	}

	return orders, nil
}
