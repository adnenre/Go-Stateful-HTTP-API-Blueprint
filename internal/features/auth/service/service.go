// internal/features/auth/service/service.go
package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"rest-api-blueprint/internal/auth"
	"rest-api-blueprint/internal/config"
	"rest-api-blueprint/internal/email"
	"rest-api-blueprint/internal/errors"
	"rest-api-blueprint/internal/features/auth/dto"
	"rest-api-blueprint/internal/features/auth/mapper"
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

// generateTokens creates a new access JWT and a long‑lived refresh token.
// The refresh token is stored in Redis with the user ID and expiry time.
// Returns the access token, refresh token, expiry in seconds, or an error.
func (s *authService) generateTokens(ctx context.Context, userID, role string) (accessToken, refreshToken string, expiresIn int, err error) {
	// 1. Generate access token (JWT)
	accessToken, err = auth.GenerateToken(userID, role, s.cfg.JWTSecret, s.cfg.JWTExpiry)
	if err != nil {
		return "", "", 0, err
	}
	expiresIn = int(s.cfg.JWTExpiry.Seconds())

	// 2. Generate cryptographically random refresh token (32 bytes, base64 URL encoded)
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", 0, errors.InternalError("failed to generate refresh token")
	}
	refreshToken = base64.URLEncoding.EncodeToString(b)

	// 3. Store refresh token metadata in Redis
	refreshData := struct {
		UserID    string    `json:"user_id"`
		ExpiresAt time.Time `json:"expires_at"`
	}{
		UserID:    userID,
		ExpiresAt: time.Now().UTC().Add(s.cfg.RefreshTokenExpiry),
	}
	jsonData, _ := json.Marshal(refreshData)
	key := fmt.Sprintf("refresh:%s", refreshToken)
	err = s.rdb.SetEx(ctx, key, jsonData, s.cfg.RefreshTokenExpiry).Err()
	if err != nil {
		return "", "", 0, errors.InternalError("failed to store refresh token")
	}
	return accessToken, refreshToken, expiresIn, nil
}

// Register creates a new user with status 'pending', generates an OTP,
// stores it in Redis, and sends it via email.
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
	go s.emailSender.SendOTP(ctx, email, otp) // async email sending
	return nil
}

// Login authenticates a user, returns a full login response containing
// access_token, refresh_token, and user profile (for mobile apps).
// The controller will additionally set httpOnly cookies for web browsers.
func (s *authService) Login(ctx context.Context, email, password string) (*dto.LoginResponse, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		// Do not reveal existence of email (security)
		return nil, errors.UnauthorizedError("Invalid email or password")
	}
	if user.Status != "active" {
		return nil, errors.UnauthorizedError("Account not activated")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.UnauthorizedError("Invalid email or password")
	}
	accessToken, refreshToken, expiresIn, err := s.generateTokens(ctx, user.ID, user.Role)
	if err != nil {
		return nil, err
	}
	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		TokenType:    "Bearer",
		User:         mapper.ToUserResponse(user),
	}, nil
}

// VerifyOtp validates the OTP, activates the user account, and returns
// the same token response as Login (access_token, refresh_token, user).
func (s *authService) VerifyOtp(ctx context.Context, email, otp string) (*dto.OTPResponse, error) {
	ok, err := redisstore.VerifyOtp(ctx, s.rdb, email, otp)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.UnauthorizedError("Invalid or expired OTP")
	}
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil || user == nil {
		return nil, errors.UnauthorizedError("User not found")
	}
	if user.Status != "pending" {
		return nil, errors.UnauthorizedError("Account already activated")
	}
	if err := s.repo.UpdateUserStatus(ctx, user.ID, "active"); err != nil {
		return nil, err
	}
	accessToken, refreshToken, expiresIn, err := s.generateTokens(ctx, user.ID, user.Role)
	if err != nil {
		return nil, err
	}
	return &dto.OTPResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		TokenType:    "Bearer",
		User:         mapper.ToUserResponse(user),
	}, nil
}

// RefreshAccessToken validates the provided refresh token, deletes the old one,
// and generates a new token pair (rotation). Returns a RefreshResponse containing
// the new tokens and expiry.
func (s *authService) RefreshAccessToken(ctx context.Context, refreshToken string) (*dto.RefreshResponse, error) {
	key := fmt.Sprintf("refresh:%s", refreshToken)
	val, err := s.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, errors.UnauthorizedError("Invalid or expired refresh token")
	}
	if err != nil {
		return nil, errors.InternalError("failed to validate refresh token")
	}
	var refreshData struct {
		UserID    string    `json:"user_id"`
		ExpiresAt time.Time `json:"expires_at"`
	}
	if err := json.Unmarshal([]byte(val), &refreshData); err != nil {
		return nil, errors.InternalError("invalid refresh token data")
	}
	// One‑time use: immediately delete the old refresh token
	s.rdb.Del(ctx, key)

	user, err := s.repo.FindByID(ctx, refreshData.UserID)
	if err != nil || user == nil {
		return nil, errors.UnauthorizedError("User not found")
	}
	newAccess, newRefresh, expiresIn, err := s.generateTokens(ctx, user.ID, user.Role)
	if err != nil {
		return nil, err
	}
	return &dto.RefreshResponse{
		AccessToken:  newAccess,
		RefreshToken: newRefresh,
		ExpiresIn:    expiresIn,
	}, nil
}

// RevokeAllUserTokens is a placeholder for future functionality.
// In a full implementation, it would delete all refresh tokens associated with a user.
func (s *authService) RevokeAllUserTokens(ctx context.Context, userID string) error {
	// Not implemented – can be added later if needed (e.g., logout from all devices)
	return nil
}

// GetSession returns the user profile for the currently authenticated user (by ID).
// This is used by the /auth/session endpoint.
func (s *authService) GetSession(ctx context.Context, userID string) (*dto.UserResponse, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return nil, errors.UnauthorizedError("User not found")
	}
	userResp := mapper.ToUserResponse(user)
	return &userResp, nil
}

// RequestPasswordReset generates a reset token, stores it in Redis, and sends it via email.
// Always returns nil to prevent user enumeration.
func (s *authService) RequestPasswordReset(ctx context.Context, email string) error {
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

// ConfirmPasswordReset verifies the reset token and updates the user's password.
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
