package service

import (
	"context"
	"rest-api-blueprint/internal/features/auth/model"
)

type Service interface {
	ListUsers(ctx context.Context, limit, offset int) ([]model.User, error)
	CreateUser(ctx context.Context, email, username, password, role string, avatar *string) (*model.User, error)
	GetUser(ctx context.Context, id string) (*model.User, error)
	UpdateUser(ctx context.Context, id string, email, username, role *string, password *string, avatar *string) error
	DeleteUser(ctx context.Context, id string) error
}
