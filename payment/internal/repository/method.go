package repository

import (
	"context"
	"database/sql"
	"errors"
	"shop/payment/internal/model"
)

type MethodRepository interface {
	Create(ctx context.Context, method model.Method) (model.Method, error)
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
