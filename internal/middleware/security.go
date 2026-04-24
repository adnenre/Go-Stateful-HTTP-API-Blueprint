package middleware

import (
	"fmt"
	"net/http"
)

// SecurityHeadersConfig holds configurable security header values.
type SecurityHeadersConfig struct {
	HSTSMaxAge int // seconds for Strict-Transport-Security
}

// SecurityHeaders adds common security headers to every HTTP response.
// These headers mitigate XSS, clickjacking, MIME sniffing, and enforce HTTPS.
func SecurityHeaders(cfg SecurityHeadersConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Prevent MIME type sniffing
			w.Header().Set("X-Content-Type-Options", "nosniff")

			// Prevent clickjacking (cannot embed API in iframe)
			w.Header().Set("X-Frame-Options", "DENY")

			// Enable XSS filtering (modern browsers, set to 1; mode=block)
			w.Header().Set("X-XSS-Protection", "1; mode=block")

			// Strict Transport Security (only sent over HTTPS)
			if r.TLS != nil {
				hstsValue := fmt.Sprintf("max-age=%d; includeSubDomains; preload", cfg.HSTSMaxAge)
				w.Header().Set("Strict-Transport-Security", hstsValue)
			}

			// Referrer policy: only send full referrer for same-origin requests
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// Content Security Policy (for API: deny all by default)
			w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")

			// Additional cache control for sensitive data (optional)
			w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
			w.Header().Set("Pragma", "no-cache") // HTTP/1.0 legacy

			next.ServeHTTP(w, r)
		})
	}
}
