package controller

import (
	"encoding/json"
	"net/http"
	"rest-api-blueprint/internal/config"
	"rest-api-blueprint/internal/errors"
	"rest-api-blueprint/internal/features/auth/dto"
	"rest-api-blueprint/internal/features/auth/service"
	"rest-api-blueprint/internal/middleware"
)

type AuthController struct {
	svc service.Service
	cfg *config.Config
}

func NewAuthController(svc service.Service, cfg *config.Config) *AuthController {
	return &AuthController{svc: svc, cfg: cfg}
}

// setTokenCookies sets httpOnly cookies for access and refresh tokens.
func setTokenCookies(w http.ResponseWriter, accessToken, refreshToken string, accessMaxAge, refreshMaxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   accessMaxAge,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Path:     "/api/v1/auth/refresh",
		MaxAge:   refreshMaxAge,
	})
}

func (c *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errDomain := errors.BadRequestError("Invalid request body: " + err.Error())
		errors.WriteProblem(w, r, errDomain, middleware.GetRequestID(r))
		return
	}
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
	w.WriteHeader(http.StatusAccepted)
}

func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errDomain := errors.BadRequestError("Invalid request body: " + err.Error())
		errors.WriteProblem(w, r, errDomain, middleware.GetRequestID(r))
		return
	}
	resp, err := c.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if domainErr, ok := err.(*errors.DomainError); ok {
			errors.WriteProblem(w, r, domainErr, middleware.GetRequestID(r))
		} else {
			errDomain := errors.InternalError(err.Error())
			errors.WriteProblem(w, r, errDomain, middleware.GetRequestID(r))
		}
		return
	}
	setTokenCookies(w, resp.AccessToken, resp.RefreshToken, resp.ExpiresIn, int(c.cfg.RefreshTokenExpiry.Seconds()))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (c *AuthController) VerifyOtp(w http.ResponseWriter, r *http.Request) {
	var req dto.VerifyOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errDomain := errors.BadRequestError("Invalid request body: " + err.Error())
		errors.WriteProblem(w, r, errDomain, middleware.GetRequestID(r))
		return
	}
	resp, err := c.svc.VerifyOtp(r.Context(), req.Email, req.OTP)
	if err != nil {
		if domainErr, ok := err.(*errors.DomainError); ok {
			errors.WriteProblem(w, r, domainErr, middleware.GetRequestID(r))
		} else {
			errDomain := errors.InternalError(err.Error())
			errors.WriteProblem(w, r, errDomain, middleware.GetRequestID(r))
		}
		return
	}
	setTokenCookies(w, resp.AccessToken, resp.RefreshToken, resp.ExpiresIn, int(c.cfg.RefreshTokenExpiry.Seconds()))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (c *AuthController) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		errors.WriteProblem(w, r, errors.UnauthorizedError("Missing refresh token"), middleware.GetRequestID(r))
		return
	}
	resp, err := c.svc.RefreshAccessToken(r.Context(), cookie.Value)
	if err != nil {
		if domainErr, ok := err.(*errors.DomainError); ok {
			errors.WriteProblem(w, r, domainErr, middleware.GetRequestID(r))
		} else {
			errDomain := errors.InternalError(err.Error())
			errors.WriteProblem(w, r, errDomain, middleware.GetRequestID(r))
		}
		return
	}
	setTokenCookies(w, resp.AccessToken, resp.RefreshToken, resp.ExpiresIn, int(c.cfg.RefreshTokenExpiry.Seconds()))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
func (c *AuthController) Session(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		errors.WriteProblem(w, r, errors.UnauthorizedError("No user claims"), middleware.GetRequestID(r))
		return
	}
	userResp, err := c.svc.GetSession(r.Context(), claims.UserID)
	if err != nil {

		if domainErr, ok := err.(*errors.DomainError); ok {
			errors.WriteProblem(w, r, domainErr, middleware.GetRequestID(r))
		} else {
			errDomain := errors.InternalError(err.Error())
			errors.WriteProblem(w, r, errDomain, middleware.GetRequestID(r))
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"user": userResp})
}

func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   -1,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Path:     "/api/v1/auth/refresh",
		MaxAge:   -1,
	})
	if userID, ok := r.Context().Value(middleware.UserKey).(string); ok && userID != "" {
		_ = c.svc.RevokeAllUserTokens(r.Context(), userID)
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c *AuthController) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req dto.PasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errDomain := errors.BadRequestError("Invalid request body: " + err.Error())
		errors.WriteProblem(w, r, errDomain, middleware.GetRequestID(r))
		return
	}
	_ = c.svc.RequestPasswordReset(r.Context(), req.Email)
	w.WriteHeader(http.StatusAccepted)
}

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
