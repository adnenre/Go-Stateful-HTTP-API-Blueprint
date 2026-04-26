package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration values.
// All sensitive values are required (no defaults) and are loaded from environment variables.
type Config struct {
	// Server
	ServerPort string // HTTP listen port (default 8080)

	// Database
	DatabaseURL string // PostgreSQL connection string (required)

	// Cache
	RedisURL string // Redis connection URL (required)

	// JWT authentication
	JWTSecret string        // Secret key for signing JWTs (required)
	JWTExpiry time.Duration // Token expiration (required)

	// Rate limiting
	RateLimitPerSecond int // Requests allowed per second per client (required)

	// CORS (Cross-Origin Resource Sharing)
	CORSAllowedOrigins   []string // Allowed origins (e.g., "http://localhost:3000")
	CORSAllowedMethods   []string // Allowed HTTP methods (e.g., "GET,POST")
	CORSAllowedHeaders   []string // Allowed request headers (e.g., "Content-Type,Authorization")
	CORSAllowCredentials bool     // Whether to allow credentials (cookies, auth headers)

	// HSTS (HTTP Strict Transport Security)
	HSTSMaxAge int // max-age in seconds (default 31536000)

	// Error documentation base URL (optional)
	ErrorDocsBaseURL string // e.g., "http://localhost:8080" or "https://api.example.com"

	// Email SMTP (optional – used for sending OTP and password reset emails)
	SMTPHost     string // SMTP server host (e.g., smtp.gmail.com)
	SMTPPort     int    // SMTP port (default 587)
	SMTPUser     string // SMTP auth username
	SMTPPassword string // SMTP auth password
	SMTPFrom     string // From email address

	// Public paths (exact) that bypass JWT authentication (optional)
	PublicPaths []string

	// Public prefixes that bypass JWT authentication (optional)
	PublicPrefixes []string
}

// Load reads configuration from environment variables and .env file.
// It fails fast (logs and exits) if any required variable is missing.
func Load() *Config {
	// Load .env file for local development (ignored if not found, which is fine for production)
	_ = godotenv.Load()

	serverPort := getEnv("SERVER_PORT", "8080")
	databaseURL, err := getEnvRequired("DATABASE_URL")
	if err != nil {
		slog.Error("configuration error", "error", err)
		os.Exit(1)
	}
	redisURL, err := getEnvRequired("REDIS_URL")
	if err != nil {
		slog.Error("configuration error", "error", err)
		os.Exit(1)
	}
	jwtSecret, err := getEnvRequired("JWT_SECRET")
	if err != nil {
		slog.Error("configuration error", "error", err)
		os.Exit(1)
	}
	jwtExpiry, err := getEnvDurationRequired("JWT_EXPIRY")
	if err != nil {
		slog.Error("configuration error", "error", err)
		os.Exit(1)
	}
	rateLimit, err := getEnvIntRequired("RATE_LIMIT_PER_SEC")
	if err != nil {
		slog.Error("configuration error", "error", err)
		os.Exit(1)
	}

	// CORS settings: defaults are permissive for local development.
	corsOrigins := getEnv("CORS_ALLOWED_ORIGINS", "*")
	corsMethods := getEnv("CORS_ALLOWED_METHODS", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
	corsHeaders := getEnv("CORS_ALLOWED_HEADERS", "Content-Type, Authorization, X-Request-ID")
	corsCredentials := getEnvBool("CORS_ALLOW_CREDENTIALS", false)

	// HSTS max-age (default 1 year = 31536000 seconds)
	hstsMaxAge := getEnvInt("HSTS_MAX_AGE", 31536000)

	// Error documentation base URL (optional)
	errorDocsBaseURL := getEnv("ERROR_DOCS_BASE_URL", "")

	// Email SMTP (optional)
	smtpHost := getEnv("SMTP_HOST", "")
	smtpPort := getEnvInt("SMTP_PORT", 587)
	smtpUser := getEnv("SMTP_USER", "")
	smtpPassword := getEnv("SMTP_PASSWORD", "")
	smtpFrom := getEnv("SMTP_FROM", "")

	// Public paths (exact) – no defaults; must be set via env if needed
	publicPathsStr := getEnv("PUBLIC_PATHS", "")
	var publicPaths []string
	if publicPathsStr != "" {
		publicPaths = strings.Split(publicPathsStr, ",")
		for i, p := range publicPaths {
			publicPaths[i] = strings.TrimSpace(p)
		}
	}

	// Public prefixes – no defaults; must be set via env if needed
	publicPrefixesStr := getEnv("PUBLIC_PREFIXES", "")
	var publicPrefixes []string
	if publicPrefixesStr != "" {
		publicPrefixes = strings.Split(publicPrefixesStr, ",")
		for i, p := range publicPrefixes {
			publicPrefixes[i] = strings.TrimSpace(p)
		}
	}

	return &Config{
		ServerPort:           serverPort,
		DatabaseURL:          databaseURL,
		RedisURL:             redisURL,
		JWTSecret:            jwtSecret,
		JWTExpiry:            jwtExpiry,
		RateLimitPerSecond:   rateLimit,
		CORSAllowedOrigins:   strings.Split(corsOrigins, ","),
		CORSAllowedMethods:   strings.Split(corsMethods, ","),
		CORSAllowedHeaders:   strings.Split(corsHeaders, ","),
		CORSAllowCredentials: corsCredentials,
		HSTSMaxAge:           hstsMaxAge,
		ErrorDocsBaseURL:     errorDocsBaseURL,
		SMTPHost:             smtpHost,
		SMTPPort:             smtpPort,
		SMTPUser:             smtpUser,
		SMTPPassword:         smtpPassword,
		SMTPFrom:             smtpFrom,
		PublicPaths:          publicPaths,
		PublicPrefixes:       publicPrefixes,
	}
}

// getEnv returns value or default if not set.
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// getEnvInt returns integer value or default if not set or invalid.
func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

// getEnvRequired returns error if the env var is empty.
func getEnvRequired(key string) (string, error) {
	val := os.Getenv(key)
	if val == "" {
		return "", fmt.Errorf("environment variable %s is required", key)
	}
	return val, nil
}

// getEnvDurationRequired parses a duration from env; error if empty or invalid.
func getEnvDurationRequired(key string) (time.Duration, error) {
	val := os.Getenv(key)
	if val == "" {
		return 0, fmt.Errorf("environment variable %s is required", key)
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		return 0, fmt.Errorf("invalid duration for %s: %w", key, err)
	}
	return d, nil
}

// getEnvIntRequired parses an integer from env; error if empty or invalid.
func getEnvIntRequired(key string) (int, error) {
	val := os.Getenv(key)
	if val == "" {
		return 0, fmt.Errorf("environment variable %s is required", key)
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("invalid integer for %s: %w", key, err)
	}
	return i, nil
}

// getEnvBool parses a boolean from env; defaults to defaultVal if not set or invalid.
func getEnvBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		b, err := strconv.ParseBool(val)
		if err == nil {
			return b
		}
	}
	return defaultVal
}
