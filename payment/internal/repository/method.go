package repository

import (
	"context"
	"database/sql"
	"errors"
	"shop/payment/internal/model"
)

type MethodRepository interface {
	Create(ctx context.Context, method model.Method) (model.Method, error)
	FindByID(ctx context.Context, id string) (model.Method, error)
}

type PostgresMethodRepository struct{}

func NewPostgresMethodRepository() *PostgresMethodRepository {
	return &PostgresMethodRepository{}
}

func (m *PostgresMethodRepository) Create(ctx context.Context, method model.Method) (model.Method, error) {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return model.Method{}, errors.New("transaction not found in context")
	}

	query := `INSERT INTO methods (id, user_id, gateway, payment_type, token) VALUES ($1, $2, $3, $4, $5)`
	_, err := tx.ExecContext(ctx, query, method.ID, method.UserID, method.Gateway, method.PaymentType, method.Token)
	if err != nil {
		return model.Method{}, err
	}

	return method, nil
}

func (m *PostgresMethodRepository) FindByID(ctx context.Context, id string) (model.Method, error) {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return model.Method{}, errors.New("transaction not found in context")
	}

	var method model.Method

	query := `SELECT id, user_id, gateway, payment_type, token FROM methods WHERE id = $1`
	err := tx.QueryRowContext(ctx, query, id).Scan(
		&method.ID,
		&method.UserID,
		&method.Gateway,
		&method.PaymentType,
		&method.Token,
	)
	if err != nil {
		return model.Method{}, err
	}

	return method, nil
}
