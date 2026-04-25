package repository

import (
	"context"
	"errors"
	"rest-api-blueprint/internal/features/auth/model"
	userModel "rest-api-blueprint/internal/features/user/model"

	"gorm.io/gorm"
)

type gormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

// FindUserByID retrieves a user by ID.
func (r *gormRepository) FindUserByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

// UpdateUser updates the user record (e.g., avatar, username).
func (r *gormRepository) UpdateUser(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// GetPreferences retrieves preferences for a user, creating default if not exists.
func (r *gormRepository) GetPreferences(ctx context.Context, userID string) (*userModel.UserPreferences, error) {
	var prefs userModel.UserPreferences
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&prefs).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create default preferences
		defaultPrefs := &userModel.UserPreferences{
			UserID:        userID,
			Notifications: boolPtr(true),
			Language:      stringPtr("en"),
		}
		if err := r.db.Create(defaultPrefs).Error; err != nil {
			return nil, err
		}
		return defaultPrefs, nil
	}
	return &prefs, err
}

// UpdatePreferences updates the preferences record.
func (r *gormRepository) UpdatePreferences(ctx context.Context, prefs *userModel.UserPreferences) error {
	return r.db.WithContext(ctx).Save(prefs).Error
}

// Helper functions
func boolPtr(b bool) *bool {
	return &b
}

func stringPtr(s string) *string {
	return &s
}
