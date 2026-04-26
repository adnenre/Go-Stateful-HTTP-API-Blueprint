package email

import "context"

type Sender interface {
	SendOTP(ctx context.Context, to, otp string) error
	SendResetToken(ctx context.Context, to, token string) error
}
