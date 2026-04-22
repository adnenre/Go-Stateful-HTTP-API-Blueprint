package mapper

import (
	"rest-api-blueprint/internal/features/health/dto"
	"rest-api-blueprint/internal/features/health/model"
	"rest-api-blueprint/internal/gen"
	"time"
)

// ToHealthResponse converts service data to generated API response.
func ToHealthResponse(status string, uptime, version string, checks map[string]string) gen.HealthResponse {
	// Convert string status to generated enum
	var dataStatus gen.HealthResponseDataStatus
	switch status {
	case "healthy":
		dataStatus = gen.Healthy
	case "unhealthy":
		dataStatus = gen.Unhealthy
	default:
		dataStatus = gen.Healthy
	}

	return gen.HealthResponse{
		Status: gen.Success,
		Data: struct {
			Checks    *map[string]string           `json:"checks,omitempty"`
			Status    gen.HealthResponseDataStatus `json:"status"`
			Timestamp time.Time                    `json:"timestamp"`
			Uptime    string                       `json:"uptime"`
			Version   string                       `json:"version"`
		}{
			Status:    dataStatus,
			Timestamp: time.Now(),
			Uptime:    uptime,
			Version:   version,
			Checks:    &checks,
		},
	}
}

// ToHealthData converts model entity to service DTO (if needed).
func ToHealthData(entity *model.HealthEntity) *dto.HealthData {
	return &dto.HealthData{
		Status:    entity.Status,
		CheckedAt: entity.CheckedAt,
	}
}
