package service

import (
	"context"
	"rest-api-blueprint/internal/features/health/dto"
)

type Service interface {
	GetHealth(ctx context.Context) (*dto.HealthData, error)
}
