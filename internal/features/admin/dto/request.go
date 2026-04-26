package dto

type CreateUserRequest struct {
	Email    string  `json:"email" validate:"required,email"`
	Password string  `json:"password" validate:"required,min=6"`
	Role     string  `json:"role" validate:"required,oneof=user admin"`
	Username string  `json:"username" validate:"required"`
	Avatar   *string `json:"avatar,omitempty"`
}
type UpdateUserRequest struct {
	Email    *string `json:"email,omitempty" validate:"omitempty,email"`
	Username *string `json:"username,omitempty"`
	Role     *string `json:"role,omitempty" validate:"omitempty,oneof=user admin"`
	Password *string `json:"password,omitempty" validate:"omitempty,min=6"`
	Avatar   *string `json:"avatar,omitempty"`
}
