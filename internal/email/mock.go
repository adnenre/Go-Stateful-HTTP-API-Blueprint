package email

import (
	"context"
	"log/slog"
)

type MockSender struct{}

func (m *MockSender) SendOTP(ctx context.Context, to, otp string) error {
	slog.Info("mock email: OTP", "to", to, "otp", otp)
	return nil
}

func (m *MockSender) SendResetToken(ctx context.Context, to, token string) error {
	slog.Info("mock email: reset token", "to", to, "token", token)
	return nil
}
