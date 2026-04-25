package dto

type LoginResponse struct {
	Token string `json:"token"`
}

type UserResponse struct {
	ID        string  `json:"id"`
	Username  string  `json:"username"`
	Email     string  `json:"email"`
	Avatar    *string `json:"avatar,omitempty"`
	Role      string  `json:"role"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}
