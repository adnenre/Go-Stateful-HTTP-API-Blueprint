package middleware

import (
	"fmt"
	"net/http"
	"strings"
)

// SecurityHeadersConfig holds configurable security header values.
type SecurityHeadersConfig struct {
	HSTSMaxAge int // seconds for Strict-Transport-Security
}

// SecurityHeaders adds common security headers to every HTTP response.
// For API responses, a strict Content‑Security‑Policy is used.
// For documentation paths (/docs/, /errors/), the CSP is relaxed to allow loading of static assets and inline scripts/styles.
func SecurityHeaders(cfg SecurityHeadersConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Headers applied to all responses
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
			w.Header().Set("Pragma", "no-cache")

			// HSTS only when connected via HTTPS
			if r.TLS != nil {
				hstsValue := fmt.Sprintf("max-age=%d; includeSubDomains; preload", cfg.HSTSMaxAge)
				w.Header().Set("Strict-Transport-Security", hstsValue)
			}

			// Content‑Security‑Policy: strict for API, relaxed for documentation
			path := r.URL.Path
			if strings.HasPrefix(path, "/docs/") || strings.HasPrefix(path, "/errors/") || path == "/docs" || path == "/errors" {
				// Allow scripts, styles, images from same origin; allow inline scripts/styles and data: images (for Swagger UI)
				w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:")
			} else {
				// Strict CSP for API endpoints
				w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
			}

			next.ServeHTTP(w, r)
		})
	}
}
