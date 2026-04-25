package repository

import (
	"context"
	"rest-api-blueprint/internal/features/auth/model"
)

type Repository interface {
	ListUsers(ctx context.Context, limit, offset int) ([]model.User, error)
	CreateUser(ctx context.Context, user *model.User) error
	FindByID(ctx context.Context, id string) (*model.User, error)
	UpdateUser(ctx context.Context, user *model.User) error
	DeleteUser(ctx context.Context, id string) error
}
