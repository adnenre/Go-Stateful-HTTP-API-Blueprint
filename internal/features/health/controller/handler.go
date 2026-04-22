package controller

import (
	"encoding/json"
	"net/http"
	"rest-api-blueprint/internal/features/health/mapper"
	"rest-api-blueprint/internal/features/health/service"
)

type HealthController struct {
	svc service.Service
}

func NewHealthController(svc service.Service) *HealthController {
	return &HealthController{svc: svc}
}

// GetHealth implements gen.ServerInterface.
func (c *HealthController) GetHealth(w http.ResponseWriter, r *http.Request) {
	data, err := c.svc.GetHealth()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	resp := mapper.ToHealthResponse(data.Status, data.Uptime, data.Version, data.Checks)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
