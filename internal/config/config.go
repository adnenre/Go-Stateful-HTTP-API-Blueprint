package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
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
	cfg := &Config{
		ServerPort:         getEnv("SERVER_PORT", "8080"),
		DatabaseURL:        getEnv("DATABASE_URL", ""),
		RedisURL:           getEnv("REDIS_URL", "localhost:6379"),
		JWTSecret:          getEnv("JWT_SECRET", ""),
		JWTExpiry:          getEnvDuration("JWT_EXPIRY", 15*time.Minute),
		RateLimitPerSecond: getEnvInt("RATE_LIMIT_PER_SEC", 10),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		d, err := time.ParseDuration(val)
		if err == nil {
			return d
		}
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
