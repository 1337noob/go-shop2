package repository

import (
	"context"
	"database/sql"
	"errors"
	"shop/order/internal/model"
	"time"
)

type OrderRepository interface {
	Create(ctx context.Context, order model.Order) (model.Order, error)
	FindByID(ctx context.Context, id string) (model.Order, error)
	UpdateStatus(ctx context.Context, id string, status model.OrderStatus) error
}

type PostgresOrderRepository struct{}

func NewPostgresOrderRepository() *PostgresOrderRepository {
	return &PostgresOrderRepository{}
}

func (p *PostgresOrderRepository) Create(ctx context.Context, order model.Order) (model.Order, error) {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return model.Order{}, errors.New("transaction not found in context")
	}

	timeNow := time.Now()
	q1 := `INSERT INTO orders (id, user_id, payment_method_id, phone, email, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := tx.ExecContext(ctx, q1, order.ID, order.UserID, order.PaymentMethodID, order.Phone, order.Email, order.Status, timeNow, timeNow)
	if err != nil {
		return model.Order{}, err
	}

	for _, item := range order.Items {
		q2 := `INSERT INTO order_items (id, order_id, product_id, quantity, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`
		_, err := tx.ExecContext(ctx, q2, item.ID, item.OrderID, item.ProductID, item.Quantity, timeNow, timeNow)
		if err != nil {
			return model.Order{}, err
		}
	}

	return order, nil
}

func (p *PostgresOrderRepository) FindByID(ctx context.Context, id string) (model.Order, error) {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return model.Order{}, errors.New("transaction not found in context")
	}

	var order model.Order
	q1 := `SELECT id, user_id, payment_method_id, phone, email, status, created_at, updated_at  FROM orders WHERE id = $1`
	err := tx.QueryRowContext(ctx, q1, id).Scan(
		&order.ID,
		&order.UserID,
		&order.PaymentMethodID,
		&order.Phone,
		&order.Email,
		&order.Status,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		return model.Order{}, err
	}

	var items []model.OrderItem
	q2 := `SELECT id, order_id, product_id, quantity, created_at, updated_at  FROM order_items WHERE order_id = $1`
	rows, err := tx.QueryContext(ctx, q2, id)
	defer rows.Close()
	if err != nil {
		return model.Order{}, err
	}
	for rows.Next() {
		var i model.OrderItem
		err = rows.Scan(
			&i.ID,
			&i.OrderID,
			&i.ProductID,
			&i.Quantity,
			&i.CreatedAt,
			&i.UpdatedAt,
		)
		if err != nil {
			return model.Order{}, err
		}
		items = append(items, i)
	}

	order.Items = items

	return order, nil
}

func (p *PostgresOrderRepository) UpdateStatus(ctx context.Context, id string, status model.OrderStatus) error {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return errors.New("transaction not found in context")
	}

	timeNow := time.Now()
	q := `UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := tx.ExecContext(ctx, q, status, timeNow, id)
	if err != nil {
		return err
	}

	return nil
}
