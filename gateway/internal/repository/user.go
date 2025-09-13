package repository

import (
	"context"
	"shop/gateway/internal/model"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *model.User) error
	FindUserByEmail(ctx context.Context, email string) (*model.User, error)
	FindUserByID(ctx context.Context, id string) (*model.User, error)
}
