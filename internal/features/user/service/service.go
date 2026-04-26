package service

import (
	"context"
	"rest-api-blueprint/internal/errors"
	authModel "rest-api-blueprint/internal/features/auth/model"
	userModel "rest-api-blueprint/internal/features/user/model"
	"rest-api-blueprint/internal/features/user/repository"
)

type userService struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) Service {
	return &userService{repo: repo}
}

func (s *userService) GetProfile(ctx context.Context, userID string) (*authModel.User, error) {
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		// Use proper domain error
		return nil, errors.NotFoundError("user")
	}
	return user, nil
}

func (s *userService) GetPreferences(ctx context.Context, userID string) (*userModel.UserPreferences, error) {
	return s.repo.GetPreferences(ctx, userID)
}

func (s *userService) UpdatePreferences(ctx context.Context, userID string, notifications *bool, language *string) error {
	// Check if user exists (optional but recommended)
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.NotFoundError("user")
	}
	prefs, err := s.repo.GetPreferences(ctx, userID)
	if err != nil {
		return err
	}
	if notifications != nil {
		prefs.Notifications = notifications
	}
	if language != nil {
		prefs.Language = language
	}
	return s.repo.UpdatePreferences(ctx, prefs)
}
