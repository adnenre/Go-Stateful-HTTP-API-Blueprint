package service

import (
	"context"
)

type Service interface {
	Register(ctx context.Context, email, username, password string, avatar *string) (userID string, err error)
	Login(ctx context.Context, email, password string) (token string, err error)
}
