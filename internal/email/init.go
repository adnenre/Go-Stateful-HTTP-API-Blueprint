package email

import (
	"log/slog"
	"rest-api-blueprint/internal/config"
)

// InitSender creates an email sender based on the configuration.
// If SMTP credentials are provided, it returns an AsyncSender wrapping an SMTPSender.
// Otherwise, it returns a MockSender that logs emails to stdout.
func InitSender(cfg *config.Config) Sender {
	if cfg.SMTPHost != "" && cfg.SMTPUser != "" && cfg.SMTPPassword != "" {
		realSender := NewSMTPSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPFrom)
		slog.Info("using SMTP email sender", "host", cfg.SMTPHost)
		return NewAsyncSender(realSender, 5)
	}
	slog.Info("using mock email sender (SMTP not configured)")
	return &MockSender{}
}
