package mapper

import (
	"rest-api-blueprint/internal/features/auth/model"
	"rest-api-blueprint/internal/features/user/dto"
	userModel "rest-api-blueprint/internal/features/user/model"
)

func ToUserProfileResponse(user *model.User) dto.UserProfileResponse {
	return dto.UserProfileResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Avatar:    user.Avatar,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func ToPreferencesResponse(prefs *userModel.UserPreferences) dto.PreferencesResponse {
	resp := dto.PreferencesResponse{
		Notifications: true,
		Language:      "en",
	}
	if prefs.Notifications != nil {
		resp.Notifications = *prefs.Notifications
	}
	if prefs.Language != nil {
		resp.Language = *prefs.Language
	}
	return resp
}
