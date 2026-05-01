package service

import (
	"context"
	"rest-api-blueprint/internal/features/auth/dto"
)

type Service interface {
	Register(ctx context.Context, email, username, password string, avatar *string) error
	Login(ctx context.Context, email, password string) (*dto.LoginResponse, error)
	VerifyOtp(ctx context.Context, email, otp string) (*dto.OTPResponse, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (*dto.RefreshResponse, error)
	RevokeAllUserTokens(ctx context.Context, userID string) error
	GetSession(ctx context.Context, userID string) (*dto.UserResponse, error)
	RequestPasswordReset(ctx context.Context, email string) error
	ConfirmPasswordReset(ctx context.Context, token, newPassword string) error
}
