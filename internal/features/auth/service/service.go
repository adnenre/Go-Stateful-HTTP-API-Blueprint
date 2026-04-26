package service

import (
	"context"
	"rest-api-blueprint/internal/auth"
	"rest-api-blueprint/internal/config"
	"rest-api-blueprint/internal/email"
	"rest-api-blueprint/internal/errors"
	"rest-api-blueprint/internal/features/auth/model"
	"rest-api-blueprint/internal/features/auth/repository"
	"rest-api-blueprint/internal/redisstore"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	repo        repository.Repository
	cfg         *config.Config
	rdb         *redis.Client
	emailSender email.Sender
}

func NewService(repo repository.Repository, cfg *config.Config, rdb *redis.Client, emailSender email.Sender) Service {
	return &authService{
		repo:        repo,
		cfg:         cfg,
		rdb:         rdb,
		emailSender: emailSender,
	}
}

func (s *authService) Register(ctx context.Context, email, username, password string, avatar *string) error {
	exists, err := s.repo.EmailExists(ctx, email)
	if err != nil {
		return err
	}
	if exists {
		return errors.ConflictError("email")
	}
	exists, err = s.repo.UsernameExists(ctx, username)
	if err != nil {
		return err
	}
	if exists {
		return errors.ConflictError("username")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user := &model.User{
		Email:    email,
		Username: username,
		Password: string(hashed),
		Avatar:   avatar,
		Role:     "user",
		Status:   "pending",
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return err
	}
	otp, err := redisstore.GenerateOTP()
	if err != nil {
		return err
	}
	if err := redisstore.StoreOTP(ctx, s.rdb, email, otp); err != nil {
		return err
	}
	go s.emailSender.SendOTP(ctx, email, otp)
	return nil
}

func (s *authService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.UnauthorizedError("Invalid email or password")
	}
	if user.Status != "active" {
		return "", errors.UnauthorizedError("Account not activated")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.UnauthorizedError("Invalid email or password")
	}
	token, err := auth.GenerateToken(user.ID, user.Role, s.cfg.JWTSecret, s.cfg.JWTExpiry)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *authService) VerifyOTP(ctx context.Context, email, otp string) (string, error) {
	ok, err := redisstore.VerifyOTP(ctx, s.rdb, email, otp)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", errors.UnauthorizedError("Invalid or expired OTP")
	}
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil || user == nil {
		return "", errors.UnauthorizedError("User not found")
	}
	if user.Status != "pending" {
		return "", errors.UnauthorizedError("Account already activated")
	}
	if err := s.repo.UpdateUserStatus(ctx, user.ID, "active"); err != nil {
		return "", err
	}
	token, err := auth.GenerateToken(user.ID, user.Role, s.cfg.JWTSecret, s.cfg.JWTExpiry)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *authService) RequestPasswordReset(ctx context.Context, email string) error {
	// Always return nil to prevent user enumeration
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil || user == nil {
		return nil
	}
	token, err := redisstore.GenerateResetToken()
	if err != nil {
		return err
	}
	if err := redisstore.StoreResetToken(ctx, s.rdb, token, user.ID); err != nil {
		return err
	}
	go s.emailSender.SendResetToken(ctx, email, token)
	return nil
}

func (s *authService) ConfirmPasswordReset(ctx context.Context, token, newPassword string) error {
	userID, err := redisstore.GetUserIDFromResetToken(ctx, s.rdb, token)
	if err != nil || userID == "" {
		return errors.UnauthorizedError("Invalid or expired reset token")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.repo.UpdateUserPassword(ctx, userID, string(hashed)); err != nil {
		return err
	}
	redisstore.DeleteResetToken(ctx, s.rdb, token)
	return nil
}
