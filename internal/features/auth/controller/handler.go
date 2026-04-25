package controller

import (
	"encoding/json"
	"net/http"
	"rest-api-blueprint/internal/errors"
	"rest-api-blueprint/internal/features/auth/dto"
	"rest-api-blueprint/internal/features/auth/service"
	"rest-api-blueprint/internal/middleware"
)

type AuthController struct {
	svc service.Service
}

func NewAuthController(svc service.Service) *AuthController {
	return &AuthController{svc: svc}
}

func (c *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteProblemSimple(w, r, http.StatusBadRequest, "Invalid request body", err.Error(), middleware.GetRequestID(r))
		return
	}
	// Basic validation
	if req.Email == "" || req.Password == "" || req.Username == "" {
		errors.WriteProblemSimple(w, r, http.StatusBadRequest, "Validation failed", "email, password, and username are required", middleware.GetRequestID(r))
		return
	}
	userID, err := c.svc.Register(r.Context(), req.Email, req.Username, req.Password, req.Avatar)
	if err != nil {
		switch err.Error() {
		case "email already exists", "username already taken":
			errors.WriteProblemSimple(w, r, http.StatusConflict, "Conflict", err.Error(), middleware.GetRequestID(r))
		default:
			errors.WriteProblemSimple(w, r, http.StatusInternalServerError, "Registration failed", err.Error(), middleware.GetRequestID(r))
		}
		return
	}
	_ = userID
	w.WriteHeader(http.StatusCreated)
	// Optionally return the created user's ID or profile (not required by OpenAPI)
}

func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteProblemSimple(w, r, http.StatusBadRequest, "Invalid request body", err.Error(), middleware.GetRequestID(r))
		return
	}
	token, err := c.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		errors.WriteProblemSimple(w, r, http.StatusUnauthorized, "Authentication failed", "Invalid email or password", middleware.GetRequestID(r))
		return
	}
	resp := dto.LoginResponse{Token: token}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
