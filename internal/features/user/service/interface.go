package service

import (
	"context"
	authModel "rest-api-blueprint/internal/features/auth/model"
	userModel "rest-api-blueprint/internal/features/user/model"
)

type Service interface {
	GetProfile(ctx context.Context, userID string) (*authModel.User, error)
	UpdatePreferences(ctx context.Context, userID string, notifications *bool, language *string) error
	GetPreferences(ctx context.Context, userID string) (*userModel.UserPreferences, error)
}
