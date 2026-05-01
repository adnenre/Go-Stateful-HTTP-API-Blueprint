package auth

import (
	"net/http"
	"rest-api-blueprint/internal/features/auth/dto"
)

var routeDTOMap = map[string]func() any{
	"POST /api/v1/auth/register":               func() any { return &dto.RegisterRequest{} },
	"POST /api/v1/auth/login":                  func() any { return &dto.LoginRequest{} },
	"POST /api/v1/auth/verify-otp":             func() any { return &dto.VerifyOTPRequest{} },
	"POST /api/v1/auth/password-reset/request": func() any { return &dto.PasswordResetRequest{} },
	"POST /api/v1/auth/password-reset/confirm": func() any { return &dto.PasswordResetConfirm{} },
}

// Resolver returns a new instance of the DTO for the given request,
// and a boolean indicating whether the route is known.
func Resolver(r *http.Request) (any, bool) {
	key := r.Method + " " + r.URL.Path
	if factory, ok := routeDTOMap[key]; ok {
		return factory(), true
	}
	return nil, false
}
