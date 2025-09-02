package repository

import (
	"context"
	"database/sql"
	"errors"
	"shop/payment/internal/model"
	"time"
)

type PaymentRepository interface {
	Create(ctx context.Context, payment model.Payment) (model.Payment, error)
	UpdateStatus(ctx context.Context, id string, status model.PaymentStatus) error
	UpdateExternalID(ctx context.Context, id string, externalID string) error
}

type PostgresPaymentRepository struct{}

func NewPostgresPaymentRepository() *PostgresPaymentRepository {
	return &PostgresPaymentRepository{}
}

func (p *PostgresPaymentRepository) Create(ctx context.Context, payment model.Payment) (model.Payment, error) {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return model.Payment{}, errors.New("transaction not found in context")
	}

	query := `INSERT INTO payments (id, order_id, user_id, amount, external_id, status, method_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := tx.ExecContext(ctx, query, payment.ID, payment.OrderID, payment.UserID, payment.Amount, nil, model.PaymentStatusPending, payment.MethodID, time.Now(), time.Now())
	if err != nil {
		return model.Payment{}, err
	}

	return payment, nil
}

func (p *PostgresPaymentRepository) UpdateStatus(ctx context.Context, id string, status model.PaymentStatus) error {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return errors.New("transaction not found in context")
	}

	query := `UPDATE payments SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := tx.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return err
	}

	return nil
}

func (p *PostgresPaymentRepository) UpdateExternalID(ctx context.Context, id string, externalID string) error {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return errors.New("transaction not found in context")
	}

	query := `UPDATE payments SET external_id = $1, updated_at = $2 WHERE id = $3`
	_, err := tx.ExecContext(ctx, query, externalID, time.Now(), id)
	if err != nil {
		return err
	}

	return nil
}
