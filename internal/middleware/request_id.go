package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type ctxKeyRequestID struct{}

// RequestIDMiddleware generates a unique ID for each request and stores it in the context.
// If the request has an X-Request-ID header, it uses that instead.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		// Set the response header so the client can see it
		w.Header().Set("X-Request-ID", requestID)

		// Store request ID in context
		ctx := context.WithValue(r.Context(), ctxKeyRequestID{}, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID extracts the request ID from the context.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(ctxKeyRequestID{}).(string); ok {
		return id
	}
	return ""
}
