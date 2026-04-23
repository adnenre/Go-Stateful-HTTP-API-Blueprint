package service

import (
	"context"
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

func (s *service) GetHealth(ctx context.Context) (*dto.HealthData, error) {
	checks := make(map[string]string)

	// Check database connectivity
	if err := s.repo.Ping(ctx); err != nil {
		checks["database"] = err.Error()
	} else {
		checks["database"] = "ok"
	}

	// Determine overall status
	status := "healthy"
	for _, v := range checks {
		if v != "ok" {
			status = "unhealthy"
			break
		}
	}

	return &dto.HealthData{
		Status:    status,
		Timestamp: time.Now().Format(time.RFC3339),
		Uptime:    time.Since(startTime).Round(time.Second).String(),
		Version:   "1.0.0",
		Checks:    checks,
	}, nil
}
