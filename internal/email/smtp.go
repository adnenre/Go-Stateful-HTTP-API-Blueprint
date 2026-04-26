package email

import (
	"context"
	"fmt"
	"net/smtp"
)

type SMTPSender struct {
	host string
	port int
	user string
	pass string
	from string
}

func NewSMTPSender(host string, port int, user, pass, from string) *SMTPSender {
	return &SMTPSender{
		host: host,
		port: port,
		user: user,
		pass: pass,
		from: from,
	}
}

func (s *SMTPSender) SendOTP(ctx context.Context, to, otp string) error {
	subject := "Your OTP Code"
	body := fmt.Sprintf("Your OTP code is: %s", otp)
	return s.send(to, subject, body)
}

func (s *SMTPSender) SendResetToken(ctx context.Context, to, token string) error {
	subject := "Password Reset"
	body := fmt.Sprintf("Use this token to reset your password: %s", token)
	return s.send(to, subject, body)
}

func (s *SMTPSender) send(to, subject, body string) error {
	auth := smtp.PlainAuth("", s.user, s.pass, s.host)
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", s.from, to, subject, body)
	return smtp.SendMail(addr, auth, s.from, []string{to}, []byte(msg))
}
