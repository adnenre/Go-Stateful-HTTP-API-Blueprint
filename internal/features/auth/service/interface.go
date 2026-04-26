package service

import "context"

type Service interface {
	Register(ctx context.Context, email, username, password string, avatar *string) error
	Login(ctx context.Context, email, password string) (string, error)
	VerifyOTP(ctx context.Context, email, otp string) (string, error)
	RequestPasswordReset(ctx context.Context, email string) error
	ConfirmPasswordReset(ctx context.Context, token, newPassword string) error
}
