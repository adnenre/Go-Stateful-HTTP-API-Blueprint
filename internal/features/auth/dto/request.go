package dto

type RegisterRequest struct {
	Email    string  `json:"email"`
	Password string  `json:"password"`
	Username string  `json:"username"`
	Avatar   *string `json:"avatar,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
