package mapper

import (
	"rest-api-blueprint/internal/features/auth/dto"
	"rest-api-blueprint/internal/features/auth/model"
	"time"
)

func ToUserResponse(user *model.User) dto.UserResponse {
	return dto.UserResponse{
		ID:            user.ID,
		Username:      user.Username,
		Email:         user.Email,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		EmailVerified: user.EmailVerified,
		Avatar:        user.Avatar,
		Role:          user.Role,
		CreatedAt:     user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     user.UpdatedAt.Format(time.RFC3339),
	}
}
