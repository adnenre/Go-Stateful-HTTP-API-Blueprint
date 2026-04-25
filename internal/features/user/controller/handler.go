package controller

import (
	"encoding/json"
	"net/http"
	"rest-api-blueprint/internal/errors"
	"rest-api-blueprint/internal/features/user/dto"
	"rest-api-blueprint/internal/features/user/mapper"
	"rest-api-blueprint/internal/features/user/service"
	"rest-api-blueprint/internal/middleware"
)

type UserController struct {
	svc service.Service
}

func NewUserController(svc service.Service) *UserController {
	return &UserController{svc: svc}
}

func (c *UserController) GetMe(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		errors.WriteProblemSimple(w, r, http.StatusUnauthorized, "Unauthorized", "missing or invalid token", middleware.GetRequestID(r))
		return
	}
	user, err := c.svc.GetProfile(r.Context(), claims.UserID)
	if err != nil {
		errors.WriteProblemSimple(w, r, http.StatusNotFound, "Not Found", err.Error(), middleware.GetRequestID(r))
		return
	}
	resp := mapper.ToUserProfileResponse(user)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (c *UserController) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		errors.WriteProblemSimple(w, r, http.StatusUnauthorized, "Unauthorized", "missing or invalid token", middleware.GetRequestID(r))
		return
	}
	var req dto.UpdatePreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteProblemSimple(w, r, http.StatusBadRequest, "Invalid request body", err.Error(), middleware.GetRequestID(r))
		return
	}
	err := c.svc.UpdatePreferences(r.Context(), claims.UserID, req.Notifications, req.Language)
	if err != nil {
		errors.WriteProblemSimple(w, r, http.StatusInternalServerError, "Failed to update preferences", err.Error(), middleware.GetRequestID(r))
		return
	}
	// Return updated preferences
	prefs, err := c.svc.GetPreferences(r.Context(), claims.UserID)
	if err != nil {
		errors.WriteProblemSimple(w, r, http.StatusInternalServerError, "Failed to retrieve preferences", err.Error(), middleware.GetRequestID(r))
		return
	}
	resp := mapper.ToPreferencesResponse(prefs)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
