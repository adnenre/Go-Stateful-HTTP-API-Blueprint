package dto

type CreateUserRequest struct {
	Email    string  `json:"email"`
	Password string  `json:"password"`
	Role     string  `json:"role"`
	Username string  `json:"username"`
	Avatar   *string `json:"avatar,omitempty"`
}

type UpdateUserRequest struct {
	Email    *string `json:"email,omitempty"`
	Username *string `json:"username,omitempty"`
	Role     *string `json:"role,omitempty"`
	Password *string `json:"password,omitempty"`
	Avatar   *string `json:"avatar,omitempty"`
}
