package service

import (
	"rest-api-blueprint/internal/features/health/dto"
	"rest-api-blueprint/internal/features/health/repository"
	"time"
)

var startTime = time.Now()

type service struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetHealth() (*dto.HealthData, error) {
	// Here you could call s.repo.Ping(ctx) to check DB health.
	// For now, assume healthy.
	return &dto.HealthData{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Uptime:    time.Since(startTime).Round(time.Second).String(),
		Version:   "dev",
		Checks:    nil, // no dependencies to check
	}, nil
}
