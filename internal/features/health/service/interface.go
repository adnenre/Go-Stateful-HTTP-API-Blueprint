package service

import "rest-api-blueprint/internal/features/health/dto"

type Service interface {
	GetHealth() (*dto.HealthData, error)
}
