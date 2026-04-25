package repository

import (
	"context"
	authModel "rest-api-blueprint/internal/features/auth/model"
	userModel "rest-api-blueprint/internal/features/user/model"
)

type Repository interface {
	FindUserByID(ctx context.Context, id string) (*authModel.User, error)
	UpdateUser(ctx context.Context, user *authModel.User) error
	GetPreferences(ctx context.Context, userID string) (*userModel.UserPreferences, error)
	UpdatePreferences(ctx context.Context, prefs *userModel.UserPreferences) error
}
