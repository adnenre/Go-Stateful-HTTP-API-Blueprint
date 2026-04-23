package middleware

import (
	"net"
	"net/http"
	"strconv"
	"strings"

	"log/slog"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

// RateLimiter wraps a Redis client and a limit.
type RateLimiter struct {
	limiter *redis_rate.Limiter
	limit   int  // requests per second
	burst   int  // burst size (optional, set same as limit for no burst)
	bypass  bool // set to true to disable rate limiting (e.g., for testing)
}

// NewRateLimiter creates a rate limiter using the given Redis client.
// limit is the number of requests allowed per second.
func NewRateLimiter(rdb *redis.Client, limit int) *RateLimiter {
	return &RateLimiter{
		limiter: redis_rate.NewLimiter(rdb),
		limit:   limit,
		burst:   limit, // no burst beyond the per-second rate
	}
}

// Middleware returns a net/http middleware that limits requests per client key.
// keyFunc extracts a unique client identifier (e.g., API key or IP) from the request.
func (rl *RateLimiter) Middleware(keyFunc func(r *http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if rl.bypass || rl.limit <= 0 {
				slog.Debug("rate limiter bypassed", "limit", rl.limit, "bypass", rl.bypass)
				next.ServeHTTP(w, r)
				return
			}

			key := keyFunc(r)
			if key == "" {
				key = "unknown"
			}

			// Use a sliding window rate limiter: allow 'limit' requests per second.
			res, err := rl.limiter.Allow(r.Context(), key, redis_rate.PerSecond(rl.limit))
			if err != nil {
				slog.Error("rate limiter redis error", "error", err, "key", key)
				http.Error(w, "rate limiter error", http.StatusInternalServerError)
				return
			}

			slog.Info("rate limit decision",
				"key", key,
				"limit", rl.limit,
				"allowed", res.Allowed,
				"remaining", res.Remaining,
				"retry_after_sec", res.RetryAfter.Seconds(),
			)

			if res.Allowed == 0 {
				w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.limit))
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("X-RateLimit-Reset", strconv.Itoa(int(res.RetryAfter.Seconds())))
				w.Header().Set("Retry-After", strconv.Itoa(int(res.RetryAfter.Seconds())))
				http.Error(w, "too many requests", http.StatusTooManyRequests)
				return
			}

			// Optional headers to inform client
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(res.Remaining))

			next.ServeHTTP(w, r)
		})
	}
}

// DefaultIPKeyFunc extracts the client IP address, normalising IPv6 localhost to IPv4.
// It respects X-Forwarded-For headers if present (for proxies).
func DefaultIPKeyFunc(r *http.Request) string {
	ip := ""

	// Check X-Forwarded-For header (common in proxy setups)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip = strings.TrimSpace(ips[0])
		}
	}

	// Fall back to RemoteAddr if no X-Forwarded-For
	if ip == "" {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr
		}
		ip = host
	}

	// Normalise IPv6 localhost to IPv4 localhost
	if ip == "::1" {
		ip = "127.0.0.1"
	}
	return ip
}
