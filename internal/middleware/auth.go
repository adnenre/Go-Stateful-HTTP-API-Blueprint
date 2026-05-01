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

// JWTAuthMiddleware validates the JWT token from either the access_token cookie
// or the Authorization: Bearer header. It injects claims into the request context.
// It skips authentication for exact paths and prefixes defined in the configuration.
func JWTAuthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	// Build fast lookup map for exact paths
	publicPathsMap := make(map[string]bool)
	for _, p := range cfg.PublicPaths {
		publicPathsMap[p] = true
	}
	publicPrefixes := cfg.PublicPrefixes

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Exact path match
			if publicPathsMap[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}
			// Prefix match
			for _, prefix := range publicPrefixes {
				if strings.HasPrefix(r.URL.Path, prefix) {
					next.ServeHTTP(w, r)
					return
				}
			}
			// Authentication required – try to extract token
			var tokenString string
			// 1. Try cookie (web BFF)
			cookie, err := r.Cookie("access_token")
			if err == nil {
				tokenString = cookie.Value
			} else {
				// 2. Try Authorization header (mobile)
				authHeader := r.Header.Get("Authorization")
				if authHeader != "" {
					parts := strings.SplitN(authHeader, " ", 2)
					if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
						tokenString = parts[1]
					}
				}
			}
			if tokenString == "" {
				errDomain := errors.UnauthorizedError("missing or invalid authentication")
				errors.WriteProblem(w, r, errDomain, GetRequestID(r))
				return
			}
			claims, err := auth.ValidateToken(tokenString, cfg.JWTSecret)
			if err != nil {
				errDomain := errors.UnauthorizedError("invalid or expired token")
				errors.WriteProblem(w, r, errDomain, GetRequestID(r))
				return
			}
			ctx := context.WithValue(r.Context(), UserKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserClaims extracts the authenticated user claims from the request context.
func GetUserClaims(ctx context.Context) *auth.Claims {
	if claims, ok := ctx.Value(UserKey).(*auth.Claims); ok {
		return claims
	}
	return nil
}
