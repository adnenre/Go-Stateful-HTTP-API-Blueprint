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
		errDomain := errors.BadRequestError("Invalid request body: " + err.Error())
		errors.WriteProblem(w, r, errDomain, middleware.GetRequestID(r))
		return
	}
	// Service now returns only error (no userID)
	err := c.svc.Register(r.Context(), req.Email, req.Username, req.Password, req.Avatar)
	if err != nil {
		if domainErr, ok := err.(*errors.DomainError); ok {
			errors.WriteProblem(w, r, domainErr, middleware.GetRequestID(r))
		} else {
			errDomain := errors.InternalError(err.Error())
			errors.WriteProblem(w, r, errDomain, middleware.GetRequestID(r))
		}
		return
	}
	// 202 Accepted – account created pending OTP verification
	w.WriteHeader(http.StatusAccepted)
}

func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errDomain := errors.BadRequestError("Invalid request body: " + err.Error())
		errors.WriteProblem(w, r, errDomain, middleware.GetRequestID(r))
		return
	}
	token, err := c.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if domainErr, ok := err.(*errors.DomainError); ok {
			errors.WriteProblem(w, r, domainErr, middleware.GetRequestID(r))
		} else {
			errDomain := errors.InternalError(err.Error())
			errors.WriteProblem(w, r, errDomain, middleware.GetRequestID(r))
		}
		return
	}
	resp := dto.LoginResponse{Token: token}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// VerifyOtp validates OTP and activates account, returns JWT.
func (c *AuthController) VerifyOtp(w http.ResponseWriter, r *http.Request) {
	var req dto.VerifyOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errDomain := errors.BadRequestError("Invalid request body: " + err.Error())
		errors.WriteProblem(w, r, errDomain, middleware.GetRequestID(r))
		return
	}
	token, err := c.svc.VerifyOtp(r.Context(), req.Email, req.OTP)
	if err != nil {
		if domainErr, ok := err.(*errors.DomainError); ok {
			errors.WriteProblem(w, r, domainErr, middleware.GetRequestID(r))
		} else {
			errDomain := errors.InternalError(err.Error())
			errors.WriteProblem(w, r, errDomain, middleware.GetRequestID(r))
		}
		return
	}
	resp := dto.LoginResponse{Token: token}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// RequestPasswordReset sends a reset token email (always returns 202).
func (c *AuthController) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req dto.PasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errDomain := errors.BadRequestError("Invalid request body: " + err.Error())
		errors.WriteProblem(w, r, errDomain, middleware.GetRequestID(r))
		return
	}
	// Service returns only error; we ignore it to avoid enumeration.
	_ = c.svc.RequestPasswordReset(r.Context(), req.Email)
	w.WriteHeader(http.StatusAccepted)
}

// ConfirmPasswordReset updates password with a valid reset token.
func (c *AuthController) ConfirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req dto.PasswordResetConfirm
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errDomain := errors.BadRequestError("Invalid request body: " + err.Error())
		errors.WriteProblem(w, r, errDomain, middleware.GetRequestID(r))
		return
	}
	err := c.svc.ConfirmPasswordReset(r.Context(), req.Token, req.NewPassword)
	if err != nil {
		if domainErr, ok := err.(*errors.DomainError); ok {
			errors.WriteProblem(w, r, domainErr, middleware.GetRequestID(r))
		} else {
			errDomain := errors.InternalError(err.Error())
			errors.WriteProblem(w, r, errDomain, middleware.GetRequestID(r))
		}
		return
	}
	w.WriteHeader(http.StatusOK)
}
