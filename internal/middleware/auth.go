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

// JWTAuthMiddleware validates the JWT token and injects claims into the request context.
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
			// Authentication required
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				err := errors.UnauthorizedError("missing authorization header")
				errors.WriteProblem(w, r, err, GetRequestID(r))
				return
			}
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				err := errors.UnauthorizedError("invalid authorization format")
				errors.WriteProblem(w, r, err, GetRequestID(r))
				return
			}
			tokenString := parts[1]
			claims, err := auth.ValidateToken(tokenString, cfg.JWTSecret)
			if err != nil {
				err := errors.UnauthorizedError("invalid or expired token")
				errors.WriteProblem(w, r, err, GetRequestID(r))
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
