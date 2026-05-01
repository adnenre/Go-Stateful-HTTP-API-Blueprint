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
type Config struct {
	// Server
	ServerPort string

	// Database
	DatabaseURL string

	// Cache
	RedisURL string

	// JWT authentication
	JWTSecret string
	JWTExpiry time.Duration

	// Refresh token
	RefreshTokenExpiry time.Duration

	// Rate limiting
	RateLimitPerSecond int

	// CORS
	CORSAllowedOrigins   []string
	CORSAllowedMethods   []string
	CORSAllowedHeaders   []string
	CORSAllowCredentials bool

	// HSTS
	HSTSMaxAge int

	// Error documentation base URL
	ErrorDocsBaseURL string

	// Email SMTP
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string

	// Public paths
	PublicPaths    []string
	PublicPrefixes []string
}

// Load reads configuration from environment variables and .env file.
func Load() *Config {
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
	refreshExpiry, err := getEnvDurationRequired("REFRESH_TOKEN_EXPIRY")
	if err != nil {
		slog.Error("configuration error", "error", err)
		os.Exit(1)
	}
	rateLimit, err := getEnvIntRequired("RATE_LIMIT_PER_SEC")
	if err != nil {
		slog.Error("configuration error", "error", err)
		os.Exit(1)
	}

	// CORS settings
	corsOrigins := getEnv("CORS_ALLOWED_ORIGINS", "*")
	corsMethods := getEnv("CORS_ALLOWED_METHODS", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
	corsHeaders := getEnv("CORS_ALLOWED_HEADERS", "Content-Type, Authorization, X-Request-ID")
	corsCredentials := getEnvBool("CORS_ALLOW_CREDENTIALS", false)

	// HSTS max-age
	hstsMaxAge := getEnvInt("HSTS_MAX_AGE", 31536000)

	// Error documentation base URL
	errorDocsBaseURL := getEnv("ERROR_DOCS_BASE_URL", "")

	// Email SMTP
	smtpHost := getEnv("SMTP_HOST", "")
	smtpPort := getEnvInt("SMTP_PORT", 587)
	smtpUser := getEnv("SMTP_USER", "")
	smtpPassword := getEnv("SMTP_PASSWORD", "")
	smtpFrom := getEnv("SMTP_FROM", "")

	// Public paths
	publicPathsStr := getEnv("PUBLIC_PATHS", "")
	var publicPaths []string
	if publicPathsStr != "" {
		publicPaths = strings.Split(publicPathsStr, ",")
		for i, p := range publicPaths {
			publicPaths[i] = strings.TrimSpace(p)
		}
	}

	// Public prefixes
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
		RefreshTokenExpiry:   refreshExpiry,
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

// Helper functions unchanged
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func getEnvRequired(key string) (string, error) {
	val := os.Getenv(key)
	if val == "" {
		return "", fmt.Errorf("environment variable %s is required", key)
	}
	return val, nil
}

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

func getEnvBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		b, err := strconv.ParseBool(val)
		if err == nil {
			return b
		}
	}
	return defaultVal
}
