package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

type ContextKey string

const RequestIDKey ContextKey = "request_id"

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
			slog.Info("[1] Generated new request ID", "id", requestID)
		} else {
			slog.Info("[1] Using client-provided request ID", "id", requestID)
		}
		w.Header().Set("X-Request-ID", requestID)
		r.Header.Set("X-Request-ID", requestID)
		slog.Info("[2] Set response and request headers", "responseHeader", w.Header().Get("X-Request-ID"), "requestHeader", r.Header.Get("X-Request-ID"))
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID extracts the request ID from the request.
// It first checks the context, then falls back to the header.
func GetRequestID(r *http.Request) string {
	if id, ok := r.Context().Value(RequestIDKey).(string); ok && id != "" {
		slog.Info("[3] GetRequestID: from context", "id", id)
		return id
	}
	// Fallback to header (guaranteed to be set by middleware)
	headerID := r.Header.Get("X-Request-ID")
	slog.Info("[3] GetRequestID: fallback to header", "id", headerID)
	return headerID
}
