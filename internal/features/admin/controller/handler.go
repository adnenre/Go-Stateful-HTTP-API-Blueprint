package controller

import (
	"encoding/json"
	"net/http"
	"rest-api-blueprint/internal/errors"
	"rest-api-blueprint/internal/features/admin/dto"
	"rest-api-blueprint/internal/features/admin/mapper"
	"rest-api-blueprint/internal/features/admin/service"
	"rest-api-blueprint/internal/gen"
	"rest-api-blueprint/internal/middleware"
)

type AdminController struct {
	svc service.Service
}

func NewAdminController(svc service.Service) *AdminController {
	return &AdminController{svc: svc}
}

func (c *AdminController) isAdmin(r *http.Request) bool {
	claims := middleware.GetUserClaims(r.Context())
	return claims != nil && claims.Role == "admin"
}

func (c *AdminController) ListUsers(w http.ResponseWriter, r *http.Request, params gen.ListUsersParams) {
	if !c.isAdmin(r) {
		err := errors.ForbiddenError("admin role required")
		errors.WriteProblem(w, r, err, middleware.GetRequestID(r))
		return
	}
	limit := 20
	if params.Limit != nil {
		limit = *params.Limit
	}
	offset := 0
	if params.Offset != nil {
		offset = *params.Offset
	}
	users, err := c.svc.ListUsers(r.Context(), limit, offset)
	if err != nil {
		// Unexpected error → internal server error
		errDomain := errors.InternalError(err.Error())
		errors.WriteProblem(w, r, errDomain, middleware.GetRequestID(r))
		return
	}
	resp := mapper.ToUserResponseList(users)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (c *AdminController) CreateUser(w http.ResponseWriter, r *http.Request) {
	if !c.isAdmin(r) {
		err := errors.ForbiddenError("admin role required")
		errors.WriteProblem(w, r, err, middleware.GetRequestID(r))
		return
	}
	var req dto.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Malformed JSON → 400 Bad Request
		errDomain := errors.BadRequestError("Invalid request body: " + err.Error())
		errors.WriteProblem(w, r, errDomain, middleware.GetRequestID(r))
		return
	}
	user, err := c.svc.CreateUser(r.Context(), req.Email, req.Username, req.Password, req.Role, req.Avatar)
	if err != nil {
		if domainErr, ok := err.(*errors.DomainError); ok {
			errors.WriteProblem(w, r, domainErr, middleware.GetRequestID(r))
		} else {
			// Unexpected error → internal server error
			errDomain := errors.InternalError(err.Error())
			errors.WriteProblem(w, r, errDomain, middleware.GetRequestID(r))
		}
		return
	}
	resp := mapper.ToUserResponse(user)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (c *AdminController) GetUser(w http.ResponseWriter, r *http.Request, id string) {
	if !c.isAdmin(r) {
		err := errors.ForbiddenError("admin role required")
		errors.WriteProblem(w, r, err, middleware.GetRequestID(r))
		return
	}
	user, err := c.svc.GetUser(r.Context(), id)
	if err != nil {
		if domainErr, ok := err.(*errors.DomainError); ok {
			errors.WriteProblem(w, r, domainErr, middleware.GetRequestID(r))
		} else {
			errDomain := errors.InternalError(err.Error())
			errors.WriteProblem(w, r, errDomain, middleware.GetRequestID(r))
		}
		return
	}
	if user == nil {
		err := errors.NotFoundError("user")
		errors.WriteProblem(w, r, err, middleware.GetRequestID(r))
		return
	}
	resp := mapper.ToUserResponse(user)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (c *AdminController) UpdateUser(w http.ResponseWriter, r *http.Request, id string) {
	if !c.isAdmin(r) {
		err := errors.ForbiddenError("admin role required")
		errors.WriteProblem(w, r, err, middleware.GetRequestID(r))
		return
	}
	var req dto.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errDomain := errors.BadRequestError("Invalid request body: " + err.Error())
		errors.WriteProblem(w, r, errDomain, middleware.GetRequestID(r))
		return
	}
	err := c.svc.UpdateUser(r.Context(), id, req.Email, req.Username, req.Role, req.Password, req.Avatar)
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

func (c *AdminController) DeleteUser(w http.ResponseWriter, r *http.Request, id string) {
	if !c.isAdmin(r) {
		err := errors.ForbiddenError("admin role required")
		errors.WriteProblem(w, r, err, middleware.GetRequestID(r))
		return
	}
	err := c.svc.DeleteUser(r.Context(), id)
	if err != nil {
		if domainErr, ok := err.(*errors.DomainError); ok {
			errors.WriteProblem(w, r, domainErr, middleware.GetRequestID(r))
		} else {
			errDomain := errors.InternalError(err.Error())
			errors.WriteProblem(w, r, errDomain, middleware.GetRequestID(r))
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
