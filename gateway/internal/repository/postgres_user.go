package repository

import (
	"context"
	"database/sql"
	"shop/gateway/internal/model"
)

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) CreateUser(ctx context.Context, user *model.User) error {
	q := `INSERT INTO users (id, name, email, password) VALUES ($1, $2, $3, $4)`
	_, err := r.db.ExecContext(ctx, q, user.ID, user.Name, user.Email, user.Password)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgresUserRepository) FindUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	q := `SELECT id, name, email, password  FROM users WHERE email = $1`
	err := r.db.QueryRowContext(ctx, q, email).Scan(
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

func (r *PostgresUserRepository) FindUserByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	q := `SELECT id, name, email, password  FROM users WHERE id = $1`
	err := r.db.QueryRowContext(ctx, q, id).Scan(
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
