package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"shop/order_saga/internal/model"
	"time"
)

type Repository interface {
	Create(ctx context.Context, saga *model.Saga) error
	Update(ctx context.Context, saga *model.Saga) error
	Find(ctx context.Context, id string) (*model.Saga, error)
}

type PostgresSagaRepo struct{}

func NewPostgresSagaRepo() *PostgresSagaRepo {
	return &PostgresSagaRepo{}
}

func (r *PostgresSagaRepo) Create(ctx context.Context, saga *model.Saga) error {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return errors.New("transaction not found in context")
	}

	stepsJSON, _ := json.Marshal(saga.Steps)
	payloadJSON, _ := json.Marshal(saga.Payload)
	createdAt := time.Now()
	_, err := tx.ExecContext(
		ctx,
		"INSERT INTO sagas (id, current_step, status, steps, payload, compensating, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		saga.ID, saga.CurrentStep, saga.Status, stepsJSON, payloadJSON, saga.Compensating, createdAt, createdAt,
	)

	return err
}

func (r *PostgresSagaRepo) Update(ctx context.Context, saga *model.Saga) error {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return errors.New("transaction not found in context")
	}

	stepsJSON, _ := json.Marshal(saga.Steps)
	payloadJSON, _ := json.Marshal(saga.Payload)
	updatedAt := time.Now()
	_, err := tx.ExecContext(
		ctx,
		"UPDATE sagas SET current_step = $1, status = $2, steps = $3, payload = $4, compensating = $5, updated_at = $6 WHERE id = $7",
		saga.CurrentStep, saga.Status, stepsJSON, payloadJSON, saga.Compensating, updatedAt, saga.ID,
	)

	return err
}

func (r *PostgresSagaRepo) Find(ctx context.Context, id string) (*model.Saga, error) {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return nil, errors.New("transaction not found in context")
	}

	var saga model.Saga
	var stepsJSON []byte
	var payloadJSON []byte
	err := tx.QueryRowContext(
		ctx,
		"SELECT id, current_step, status, steps, payload, compensating, created_at FROM sagas WHERE id = $1",
		id,
	).Scan(&saga.ID, &saga.CurrentStep, &saga.Status, &stepsJSON, &payloadJSON, &saga.Compensating, &saga.CreatedAt)

	json.Unmarshal(stepsJSON, &saga.Steps)
	json.Unmarshal(payloadJSON, &saga.Payload)

	return &saga, err
}
