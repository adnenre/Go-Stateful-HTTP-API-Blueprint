package controller

import (
	"encoding/json"
	"net/http"
	"rest-api-blueprint/internal/errors"
	"rest-api-blueprint/internal/features/health/mapper"
	"rest-api-blueprint/internal/features/health/service"
	"rest-api-blueprint/internal/middleware"
)

type HealthController struct {
	svc service.Service
}

func NewHealthController(svc service.Service) *HealthController {
	return &HealthController{svc: svc}
}

// GetHealth implements gen.ServerInterface.
func (c *HealthController) GetHealth(w http.ResponseWriter, r *http.Request) {
	data, err := c.svc.GetHealth(r.Context())
	if err != nil {
		instance := middleware.GetRequestID(r)
		errDomain := errors.InternalError("Failed to check health: " + err.Error())
		errors.WriteProblem(w, r, errDomain, instance)
		return
	}
	resp := mapper.ToHealthResponse(data.Status, data.Uptime, data.Version, data.Checks)
	w.Header().Set("Content-Type", "application/json")
	if data.Status != "healthy" {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	json.NewEncoder(w).Encode(resp)
}
