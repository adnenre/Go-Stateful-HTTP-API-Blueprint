package service

import (
	"context"
	"errors"
	"rest-api-blueprint/internal/features/admin/repository"
	"rest-api-blueprint/internal/features/auth/model"

	"golang.org/x/crypto/bcrypt"
)

type adminService struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) Service {
	return &adminService{repo: repo}
}

func (s *adminService) ListUsers(ctx context.Context, limit, offset int) ([]model.User, error) {
	return s.repo.ListUsers(ctx, limit, offset)
}

func (s *adminService) CreateUser(ctx context.Context, email, username, password, role string, avatar *string) (*model.User, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &model.User{
		Email:    email,
		Username: username,
		Password: string(hashed),
		Role:     role,
		Avatar:   avatar,
	}
	err = s.repo.CreateUser(ctx, user)
	return user, err
}

func (s *adminService) GetUser(ctx context.Context, id string) (*model.User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *adminService) UpdateUser(ctx context.Context, id string, email, username, role *string, password *string, avatar *string) error {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}
	if email != nil {
		user.Email = *email
	}
	if username != nil {
		user.Username = *username
	}
	if role != nil {
		user.Role = *role
	}
	if password != nil && *password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user.Password = string(hashed)
	}
	if avatar != nil {
		user.Avatar = avatar
	}
	return s.repo.UpdateUser(ctx, user)
}

func (s *adminService) DeleteUser(ctx context.Context, id string) error {
	// Optional: check existence first
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	return s.repo.DeleteUser(ctx, id)
}
