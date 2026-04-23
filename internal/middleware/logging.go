package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// Logging middleware logs each request with method, path, status, latency, and request ID.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// Capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)
		latency := time.Since(start).Milliseconds()

		// Extract request ID from context
		requestID := GetRequestID(r.Context())

		slog.Info("http request",
			"request_id", requestID,
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
			"status", rw.statusCode,
			"latency_ms", latency,
		)
	})
}

// responseWriter captures the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
