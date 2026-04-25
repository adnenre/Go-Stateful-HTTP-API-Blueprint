package service

import (
	"context"
	"errors"
	"rest-api-blueprint/internal/auth"
	"rest-api-blueprint/internal/config"
	"rest-api-blueprint/internal/features/auth/model"
	"rest-api-blueprint/internal/features/auth/repository"

	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	repo repository.Repository
	cfg  *config.Config
}

func NewService(repo repository.Repository, cfg *config.Config) Service {
	return &authService{repo: repo, cfg: cfg}
}

func (s *authService) Register(ctx context.Context, email, username, password string, avatar *string) (string, error) {
	exists, err := s.repo.EmailExists(ctx, email)
	if err != nil {
		return "", err
	}
	if exists {
		return "", errors.New("email already exists")
	}
	exists, err = s.repo.UsernameExists(ctx, username)
	if err != nil {
		return "", err
	}
	if exists {
		return "", errors.New("username already taken")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	user := &model.User{
		Email:    email,
		Username: username,
		Password: string(hashed),
		Avatar:   avatar,
		Role:     "user",
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return "", err
	}
	return user.ID, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}
	token, err := auth.GenerateToken(user.ID, user.Role, s.cfg.JWTSecret, s.cfg.JWTExpiry)
	if err != nil {
		return "", err
	}
	return token, nil
}
