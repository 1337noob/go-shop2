package repository

import (
	"context"
	"database/sql"
	"errors"
	"shop/gateway/internal/model"
)

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (repo *PostgresUserRepository) CreateUser(ctx context.Context, user *model.User) error {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return errors.New("transaction not found in context")
	}

	q := `INSERT INTO users (id, name, email, password) VALUES ($1, $2, $3, $4)`
	_, err := tx.ExecContext(ctx, q, user.ID, user.Name, user.Email, user.Password)
	if err != nil {
		return err
	}

	return nil
}

func (repo *PostgresUserRepository) FindUserByEmail(ctx context.Context, email string) (*model.User, error) {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return nil, errors.New("transaction not found in context")
	}

	var user model.User
	q := `SELECT id, name, email, password  FROM users WHERE email = $1`
	err := tx.QueryRowContext(ctx, q, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (repo *PostgresUserRepository) FindUserByID(ctx context.Context, id string) (*model.User, error) {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return nil, errors.New("transaction not found in context")
	}

	var user model.User
	q := `SELECT id, name, email, password  FROM users WHERE id = $1`
	err := tx.QueryRowContext(ctx, q, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
