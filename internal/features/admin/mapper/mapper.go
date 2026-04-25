package mapper

import (
	"rest-api-blueprint/internal/features/admin/dto"
	"rest-api-blueprint/internal/features/auth/model"
)

func ToUserResponse(user *model.User) dto.UserResponse {
	return dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Avatar:    user.Avatar,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func ToUserResponseList(users []model.User) []dto.UserResponse {
	resp := make([]dto.UserResponse, len(users))
	for i, u := range users {
		resp[i] = ToUserResponse(&u)
	}
	return resp
}
