package middleware

import (
	"log/slog"
	"net/http"
	"rest-api-blueprint/internal/errors"
)

func PanicRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic recovered", "panic", rec, "path", r.URL.Path)
				err := errors.InternalError("An unexpected error occurred")
				errors.WriteProblem(w, r, err, GetRequestID(r))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
