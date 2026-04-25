package middleware

import (
	"context"
	"net/http"
	"strings"

	"rest-api-blueprint/internal/auth"
	"rest-api-blueprint/internal/config"
	"rest-api-blueprint/internal/errors"
)

type ContextKeyUser struct{}

var UserKey = ContextKeyUser{}

func JWTAuthMiddleware(cfg *config.Config, publicPaths map[string]bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if publicPaths[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				errors.WriteProblemSimple(w, r, http.StatusUnauthorized, "Unauthorized", "missing authorization header", GetRequestID(r))
				return
			}
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				errors.WriteProblemSimple(w, r, http.StatusUnauthorized, "Unauthorized", "invalid authorization format", GetRequestID(r))
				return
			}
			tokenString := parts[1]
			claims, err := auth.ValidateToken(tokenString, cfg.JWTSecret)
			if err != nil {
				errors.WriteProblemSimple(w, r, http.StatusUnauthorized, "Unauthorized", "invalid or expired token", GetRequestID(r))
				return
			}
			ctx := context.WithValue(r.Context(), UserKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserClaims(ctx context.Context) *auth.Claims {
	if claims, ok := ctx.Value(UserKey).(*auth.Claims); ok {
		return claims
	}
	return nil
}
