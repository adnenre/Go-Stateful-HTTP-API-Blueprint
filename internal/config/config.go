package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort         string
	DatabaseURL        string
	RedisURL           string
	JWTSecret          string
	JWTExpiry          time.Duration
	RateLimitPerSecond int
}

func Load() (*Config, error) {
	// Load .env file for local development (ignore if not present)
	_ = godotenv.Load()

	serverPort := getEnv("SERVER_PORT", "8080")
	databaseURL, err := getEnvRequired("DATABASE_URL")
	if err != nil {
		return nil, err
	}
	redisURL, err := getEnvRequired("REDIS_URL")
	if err != nil {
		return nil, err
	}
	jwtSecret, err := getEnvRequired("JWT_SECRET")
	if err != nil {
		return nil, err
	}
	jwtExpiry, err := getEnvDurationRequired("JWT_EXPIRY")
	if err != nil {
		return nil, err
	}
	rateLimit, err := getEnvIntRequired("RATE_LIMIT_PER_SEC")
	if err != nil {
		return nil, err
	}

	return &Config{
		ServerPort:         serverPort,
		DatabaseURL:        databaseURL,
		RedisURL:           redisURL,
		JWTSecret:          jwtSecret,
		JWTExpiry:          jwtExpiry,
		RateLimitPerSecond: rateLimit,
	}, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
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
