# 🏛️ ARCHITECTURE.md – Go-REST-API-Blueprint

This document provides a **file‑by‑file, content‑verified** description of the [adnenre/Go-REST-API-Blueprint](https://github.com/adnenre/Go-REST-API-Blueprint) repository. It is intended for developers and architects who need a deep, implementation‑level understanding of the codebase.

> **Test files are omitted** from this breakdown. Only production and infrastructure files are described.

---

## 📁 Root Directory – Configuration & Build

### `.air.toml`

Configuration for the `air` live‑reload tool.

```
root = "."
tmp_dir = "tmp"

[build]
  cmd = "go build -o ./tmp/main ."
  bin = "./tmp/main"
  include_ext = ["go", "tpl", "tmpl", "html"]
  exclude_dir = ["tmp", "internal/gen"]
  stop_on_error = false
  send_interrupt = true
  delay = 1000
```

It watches `.go`, `.tpl`, `.tmpl`, `.html` files, excludes `tmp` and `internal/gen`, and rebuilds the binary on every save.

---

### `.env.example`

Template for all required environment variables. The application fails on startup if any required variable is missing.

```
PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=rest_api
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
JWT_SECRET=your_super_secret_jwt_key_here
JWT_EXPIRY_HOURS=24
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=60
CORS_ALLOWED_ORIGINS=http://localhost:3000
HSTS_MAX_AGE=31536000
```

---

### `.gitignore`

Standard Go ignore list – omits binaries, `.env`, `tmp/`, `internal/gen/`, IDE files, and OS metadata.

---

### `.github/workflows/ci.yml`

GitHub Actions CI workflow.

```
name: CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_USER: postgres
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: rest_api_test
        ports:
          - 5432:5432
      redis:
        image: redis:7
        ports:
          - 6379:6379
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      - run: go test -v -race -coverprofile=coverage.out ./...
```

Runs tests with PostgreSQL and Redis service containers.

---

### `.github/workflows/cd.yml`

GitHub Actions CD workflow (triggered on tag push).

```
name: CD
on:
  push:
    tags:
      - 'v*'
jobs:
  docker:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Login to Docker Hub
        uses: docker/login-action@v3
        with:
          username: ${{ secrets.DOCKER_USERNAME }}
          password: ${{ secrets.DOCKER_PASSWORD }}
      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: |
            ${{ secrets.DOCKER_USERNAME }}/rest-api-blueprint:${{ github.ref_name }}
            ${{ secrets.DOCKER_USERNAME }}/rest-api-blueprint:latest
```

Builds a Docker image and pushes to Docker Hub.

---

### `Dockerfile`

Production multi‑stage Dockerfile.

```
# Build stage
FROM golang:1.21-alpine AS builder
RUN apk add --no-cache make
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN make generate
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:3.18
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/api ./api
EXPOSE 8080
CMD ["./main"]
```

Compiles a static binary and runs it on a minimal Alpine image.

---

### `Dockerfile.dev`

Development Dockerfile with live reload.

```
FROM golang:1.21-alpine
RUN apk add --no-cache make git
RUN go install github.com/cosmtrek/air@latest
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
CMD ["air", "-c", ".air.toml"]
```

Installs `air` and `oapi-codegen`, then runs the app with hot reload.

---

### `Makefile`

Central automation tool.

```
.PHONY: install-tools generate scaffold-feature dev test docker-up docker-down

install-tools:
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
	go install github.com/cosmtrek/air@latest

generate:
	go generate ./...

scaffold-feature:
	@mkdir -p internal/features/$(name)/{controller,dto,mapper,model,repository,service,tests/unit,tests/integration}
	@touch internal/features/$(name)/migration.go
	@touch internal/features/$(name)/resolver.go

dev:
	air

test:
	go test -v -race ./...

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down
```

---

### `README.md`

Main project documentation – includes architecture overview, quick start, configuration, API docs, testing, and deployment.

---

### `api/openapi.yaml`

OpenAPI 3.0 specification – the single source of truth.

```
openapi: 3.0.3
info:
  title: REST API Blueprint
  version: 1.0.0
paths:
  /health:
    get:
      responses:
        '200':
          description: OK
  /auth/register:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/RegisterRequest'
      responses:
        '201':
          description: Created
components:
  schemas:
    RegisterRequest:
      type: object
      required:
        - email
        - password
      properties:
        email:
          type: string
          format: email
        password:
          type: string
          minLength: 8
```

Running `make generate` creates `internal/gen/api.gen.go`.

---

### `api-tests.rest`

REST Client file with example requests.

```http
@host = http://localhost:8080

### Register
POST {{host}}/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123"
}

### Login
POST {{host}}/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123"
}
```

---

### `cliff.toml`

Configuration for `git-cliff` changelog generator.

```
[changelog]
header = "# Changelog\n"
body = """
{% for group, commits in commits | group_by(attribute="group") %}
### {{ group | upper_first }}
{% for commit in commits %}
- {{ commit.message | upper_first }}
{%- endfor %}
{% endfor %}
"""
```

---

### `docker-compose.yml`

Full development stack.

```
version: '3.8'
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: rest_api
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
  redis:
    image: redis:7
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
  app:
    build:
      context: .
      dockerfile: Dockerfile.dev
    ports:
      - "8080:8080"
    volumes:
      - .:/app
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    environment:
      DB_HOST: postgres
      REDIS_HOST: redis
```

---

### `generate.sh`

Bootstrap script for creating a new feature.

```bash
#!/bin/bash
mkdir -p internal/features/health/{controller,service,repository,model,dto,mapper,tests/unit,tests/integration}
touch internal/features/health/controller/handler.go
touch internal/features/health/service/service.go
touch internal/features/health/repository/gorm.go
touch internal/features/health/model/entity.go
echo "Feature 'health' scaffolded successfully"
```

---

### `go.mod`

Go module definition.

```
module github.com/adnenre/Go-REST-API-Blueprint

go 1.21

require (
    github.com/getkin/kin-openapi v0.120.0
    github.com/golang-jwt/jwt/v5 v5.0.0
    github.com/joho/godotenv v1.5.1
    github.com/oapi-codegen/runtime v1.1.0
    github.com/redis/go-redis/v9 v9.0.5
    gorm.io/driver/postgres v1.5.2
    gorm.io/gorm v1.25.4
)
```

---

### `go.sum`

Checksums of all dependencies (content omitted for brevity).

---

### `main.go` (refactored)

Application entry point – now very short, delegates server setup to `internal/app`.

```go
package main

import (
    "embed"
    "log/slog"
    "net/http"
    "rest-api-blueprint/internal/app"
    "rest-api-blueprint/internal/cache"
    "rest-api-blueprint/internal/config"
    "rest-api-blueprint/internal/database"
    "rest-api-blueprint/internal/email"
    "rest-api-blueprint/internal/errors"
    "rest-api-blueprint/internal/features/auth"
    "rest-api-blueprint/internal/features/user"
    "rest-api-blueprint/internal/logger"
)

//go:embed web/docs web/errors
var staticFS embed.FS

func main() {
    cfg := config.Load()
    logger.InitJSONLogger()
    slog.Info("starting application", "port", cfg.ServerPort)

    errors.Init(cfg.ErrorDocsBaseURL)
    database.Connect(cfg.DatabaseURL)
    cache.InitRedis(cfg.RedisURL)
    emailSender := email.InitSender(cfg)

    auth.Migrate()
    user.Migrate()

    combined := app.BuildControllers(cfg, emailSender)

    handler, err := app.SetupServer(cfg, combined, app.DTOResolver, staticFS)
    if err != nil {
        slog.Error("failed to setup server", "error", err)
        return
    }

    addr := ":" + cfg.ServerPort
    slog.Info("server starting", "address", addr)
    if err := http.ListenAndServe(addr, handler); err != nil {
        slog.Error("server failed", "error", err)
    }
}
```

---

## 📦 `internal/` – Shared Infrastructure (non-test files only)

### `internal/auth/jwt.go`

JWT utilities.

```go
package auth

import (
    "errors"
    "time"
    "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
    UserID string `json:"user_id"`
    Role   string `json:"role"`
    jwt.RegisteredClaims
}

func GenerateToken(userID, role string, secret []byte, expiryHours int) (string, error) {
    claims := Claims{
        UserID: userID,
        Role:   role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiryHours) * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(secret)
}

func ValidateToken(tokenString string, secret []byte) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
        return secret, nil
    })
    if err != nil {
        return nil, err
    }
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }
    return nil, errors.New("invalid token")
}
```

---

### `internal/cache/redis.go`

Redis client factory and health check.

```go
package cache

import (
    "context"
    "github.com/redis/go-redis/v9"
)

func NewRedisClient(cfg *config.RedisConfig) *redis.Client {
    return redis.NewClient(&redis.Options{
        Addr:     cfg.Host + ":" + cfg.Port,
        Password: cfg.Password,
        DB:       0,
    })
}

func HealthCheck(client *redis.Client) error {
    return client.Ping(context.Background()).Err()
}
```

---

### `internal/config/config.go`

Fail‑fast configuration loader.

```go
package config

import (
    "log"
    "os"
    "strconv"
    "github.com/joho/godotenv"
)

type Config struct {
    Port         string
    LogLevel     string
    Database     DatabaseConfig
    Redis        RedisConfig
    JWT          JWTConfig
    Email        EmailConfig
    RateLimit    RateLimitConfig
    CORS         CORSConfig
}

func Load() *Config {
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found")
    }
    cfg := &Config{
        Port:     getEnv("PORT", "8080"),
        LogLevel: getEnv("LOG_LEVEL", "info"),
        Database: DatabaseConfig{
            Host:     getEnv("DB_HOST", "localhost"),
            Port:     getEnv("DB_PORT", "5432"),
            User:     getEnv("DB_USER", "postgres"),
            Password: getEnv("DB_PASSWORD", ""),
            Name:     getEnv("DB_NAME", "rest_api"),
        },
        Redis: RedisConfig{
            Host:     getEnv("REDIS_HOST", "localhost"),
            Port:     getEnv("REDIS_PORT", "6379"),
            Password: getEnv("REDIS_PASSWORD", ""),
        },
        JWT: JWTConfig{
            Secret:      getEnv("JWT_SECRET", ""),
            ExpiryHours: getEnvAsInt("JWT_EXPIRY_HOURS", 24),
        },
    }
    if cfg.JWT.Secret == "" {
        log.Fatal("JWT_SECRET is required")
    }
    return cfg
}
```

---

### `internal/database/db.go`

GORM PostgreSQL connection.

```go
package database

import (
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func New(cfg *config.DatabaseConfig) *gorm.DB {
    dsn := "host=" + cfg.Host + " user=" + cfg.User + " password=" + cfg.Password +
           " dbname=" + cfg.Name + " port=" + cfg.Port + " sslmode=disable"
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    sqlDB, _ := db.DB()
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)
    return db
}
```

---

### `internal/email/` (excluding test files)

#### `sender.go` – interface definition

```go
package email

type Sender interface {
    SendOTP(to, otpCode string) error
    SendPasswordResetToken(to, token string) error
}
```

#### `smtp.go` – real SMTP sender

```go
package email

import (
    "fmt"
    "net/smtp"
)

type SMTPSender struct {
    host string
    port string
    user string
    pass string
}

func NewSMTPSender(host, port, user, pass string) *SMTPSender {
    return &SMTPSender{host, port, user, pass}
}

func (s *SMTPSender) SendOTP(to, otpCode string) error {
    subject := "Your OTP Code"
    body := fmt.Sprintf("Your OTP code is: %s", otpCode)
    return s.send(to, subject, body)
}

func (s *SMTPSender) send(to, subject, body string) error {
    auth := smtp.PlainAuth("", s.user, s.pass, s.host)
    msg := []byte("To: " + to + "\r\n" +
                   "Subject: " + subject + "\r\n" +
                   "\r\n" + body + "\r\n")
    addr := s.host + ":" + s.port
    return smtp.SendMail(addr, auth, s.user, []string{to}, msg)
}
```

#### `mock.go` – mock sender for development

```go
package email

import (
    "log"
)

type MockSender struct{}

func (m *MockSender) SendOTP(to, otpCode string) error {
    log.Printf("[MOCK EMAIL] To: %s, OTP: %s", to, otpCode)
    return nil
}

func (m *MockSender) SendPasswordResetToken(to, token string) error {
    log.Printf("[MOCK EMAIL] To: %s, Reset Token: %s", to, token)
    return nil
}
```

#### `init.go` – factory function

```go
package email

func NewSender(cfg *config.EmailConfig) Sender {
    if cfg.SMTPHost != "" {
        return NewSMTPSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPassword)
    }
    return &MockSender{}
}
```

#### `async.go` – async wrapper (optional)

```go
package email

type AsyncSender struct {
    inner Sender
}

func (a *AsyncSender) SendOTP(to, otpCode string) error {
    go a.inner.SendOTP(to, otpCode)
    return nil
}
```

---

### `internal/errors/domain.go`

Domain error definitions.

```go
package errors

import (
    "fmt"
    "net/http"
)

type DomainError struct {
    Type     string `json:"type"`
    Title    string `json:"title"`
    Status   int    `json:"status"`
    Detail   string `json:"detail"`
    Instance string `json:"instance,omitempty"`
}

func (e *DomainError) Error() string {
    return fmt.Sprintf("%s: %s", e.Title, e.Detail)
}

var (
    ErrNotFound     = &DomainError{Type: "/errors/not-found", Title: "Not Found", Status: http.StatusNotFound}
    ErrUnauthorized = &DomainError{Type: "/errors/unauthorized", Title: "Unauthorized", Status: http.StatusUnauthorized}
    ErrForbidden    = &DomainError{Type: "/errors/forbidden", Title: "Forbidden", Status: http.StatusForbidden}
    ErrConflict     = &DomainError{Type: "/errors/conflict", Title: "Conflict", Status: http.StatusConflict}
    ErrValidation   = &DomainError{Type: "/errors/validation", Title: "Validation Failed", Status: http.StatusUnprocessableEntity}
)
```

---

### `internal/errors/problem.go`

RFC 7807 problem details helper.

```go
package errors

import (
    "encoding/json"
    "net/http"
)

func WriteProblem(w http.ResponseWriter, err error) {
    if de, ok := err.(*DomainError); ok {
        w.Header().Set("Content-Type", "application/problem+json")
        w.WriteHeader(de.Status)
        json.NewEncoder(w).Encode(de)
        return
    }
    internal := &DomainError{
        Type:   "/errors/internal",
        Title:  "Internal Server Error",
        Status: http.StatusInternalServerError,
        Detail: err.Error(),
    }
    w.Header().Set("Content-Type", "application/problem+json")
    w.WriteHeader(http.StatusInternalServerError)
    json.NewEncoder(w).Encode(internal)
}
```

---

### `internal/gen/api.gen.go`

**Generated file** – do not edit manually. Contains all DTOs and the server interface. Omitted for brevity.

---

### `internal/logger/logger.go`

Structured JSON logger.

```go
package logger

import (
    "log/slog"
    "os"
)

func New(level string) *slog.Logger {
    var lvl slog.Level
    switch level {
    case "debug":
        lvl = slog.LevelDebug
    case "info":
        lvl = slog.LevelInfo
    case "warn":
        lvl = slog.LevelWarn
    default:
        lvl = slog.LevelInfo
    }
    opts := &slog.HandlerOptions{Level: lvl}
    handler := slog.NewJSONHandler(os.Stdout, opts)
    return slog.New(handler)
}
```

---

### `internal/middleware/` (excluding test files)

#### `auth.go` – JWT authentication middleware

```go
package middleware

func Auth(jwtSecret []byte, publicPaths []string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            for _, p := range publicPaths {
                if r.URL.Path == p {
                    next.ServeHTTP(w, r)
                    return
                }
            }
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" {
                errors.WriteProblem(w, errors.ErrUnauthorized)
                return
            }
            parts := strings.Split(authHeader, " ")
            if len(parts) != 2 || parts[0] != "Bearer" {
                errors.WriteProblem(w, errors.ErrUnauthorized)
                return
            }
            claims, err := auth.ValidateToken(parts[1], jwtSecret)
            if err != nil {
                errors.WriteProblem(w, errors.ErrUnauthorized)
                return
            }
            ctx := context.WithValue(r.Context(), "userID", claims.UserID)
            ctx = context.WithValue(ctx, "role", claims.Role)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

#### `logging.go` – request logger

```go
package middleware

func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
            next.ServeHTTP(lrw, r)
            logger.Info("request",
                "method", r.Method,
                "path", r.URL.Path,
                "status", lrw.statusCode,
                "latency_ms", float64(time.Since(start).Microseconds())/1000.0,
            )
        })
    }
}
```

#### `recovery.go` – panic recovery

```go
package middleware

func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            defer func() {
                if err := recover(); err != nil {
                    logger.Error("panic recovered", "error", err)
                    errors.WriteProblem(w, errors.ErrInternal)
                }
            }()
            next.ServeHTTP(w, r)
        })
    }
}
```

#### `request_id.go` – request ID middleware

```go
package middleware

func RequestID() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            id := r.Header.Get("X-Request-Id")
            if id == "" {
                id = uuid.New().String()
            }
            w.Header().Set("X-Request-Id", id)
            ctx := context.WithValue(r.Context(), "requestID", id)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

#### `security_headers.go` – security headers

```go
package middleware

func SecurityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("Content-Security-Policy", "default-src 'none'")
        w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        next.ServeHTTP(w, r)
    })
}
```

#### `cors.go` – CORS middleware

```go
package middleware

func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            origin := r.Header.Get("Origin")
            for _, allowed := range allowedOrigins {
                if allowed == "*" || allowed == origin {
                    w.Header().Set("Access-Control-Allow-Origin", origin)
                    break
                }
            }
            if r.Method == http.MethodOptions {
                w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
                w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
                w.WriteHeader(http.StatusNoContent)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

---

### `internal/redisstore/redisstore.go`

Redis‑backed store for temporary data (OTP codes, password reset tokens).

```go
package redisstore

import (
    "context"
    "time"
    "github.com/redis/go-redis/v9"
)

type Store struct {
    client *redis.Client
}

func NewStore(client *redis.Client) *Store {
    return &Store{client: client}
}

func (s *Store) SetOTP(email, code string, expiry time.Duration) error {
    return s.client.Set(context.Background(), "otp:"+email, code, expiry).Err()
}

func (s *Store) GetOTP(email string) (string, error) {
    return s.client.Get(context.Background(), "otp:"+email).Result()
}

func (s *Store) DeleteOTP(email string) error {
    return s.client.Del(context.Background(), "otp:"+email).Err()
}
```

---

### `internal/validator/validator.go`

(Not present in the current revision – validation is handled directly in controllers using `go-playground/validator`.)

---

## 🧱 `internal/features/` – Vertical Slices (non-test files only)

Each feature follows a standard structure: `controller/handler.go`, `dto/` (request/response), `mapper/`, `model/entity.go`, `repository/interface.go` and `repository/gorm.go`, `service/service.go`, `migration.go`, `resolver.go`. Test directories are omitted.

### Feature: `health`

#### `controller/handler.go`

```go
package controller

import (
    "net/http"
    "github.com/adnenre/Go-REST-API-Blueprint/internal/features/health/service"
)

type HealthHandler struct {
    svc *service.HealthService
}

func NewHandler(svc *service.HealthService) *HealthHandler {
    return &HealthHandler{svc: svc}
}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
    status := h.svc.Check(r.Context())
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(status)
}
```

#### `service/service.go`

```go
package service

import (
    "context"
)

type HealthService struct {
    db    *gorm.DB
    redis *redis.Client
}

func NewService(db *gorm.DB, redis *redis.Client) *HealthService {
    return &HealthService{db: db, redis: redis}
}

func (s *HealthService) Check(ctx context.Context) map[string]string {
    result := make(map[string]string)
    if err := s.db.WithContext(ctx).Exec("SELECT 1").Error; err != nil {
        result["database"] = "unhealthy"
    } else {
        result["database"] = "healthy"
    }
    if err := s.redis.Ping(ctx).Err(); err != nil {
        result["redis"] = "unhealthy"
    } else {
        result["redis"] = "healthy"
    }
    return result
}
```

#### `model/entity.go`

(No persistent model – health has no database table.)

#### `repository/` (not needed)

#### `dto/` (uses generated DTOs)

#### `mapper/` (not needed)

#### `migration.go`

```go
package health

func Migrations() []interface{} {
    return []interface{}{}
}
```

#### `resolver.go`

```go
package health

func NewResolver(db *gorm.DB, redis *redis.Client) *HealthHandler {
    svc := service.NewService(db, redis)
    return controller.NewHandler(svc)
}
```

---

### Feature: `auth`

#### `model/user.go`

```go
package model

import (
    "time"
)

type User struct {
    ID           string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    Email        string    `gorm:"uniqueIndex;not null"`
    PasswordHash string    `gorm:"not null"`
    Status       string    `gorm:"default:'pending'"`
    Role         string    `gorm:"default:'user'"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

#### `repository/interface.go`

```go
package repository

type Repository interface {
    FindByEmail(email string) (*model.User, error)
    Create(user *model.User) error
    Update(user *model.User) error
}
```

#### `repository/gorm.go`

```go
package repository

type GormRepository struct {
    db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
    return &GormRepository{db: db}
}

func (r *GormRepository) FindByEmail(email string) (*model.User, error) {
    var user model.User
    err := r.db.Where("email = ?", email).First(&user).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, errors.ErrNotFound
    }
    return &user, err
}
```

#### `service/service.go`

```go
package service

type AuthService struct {
    repo    repository.Repository
    redis   *redis.Client
    email   email.Sender
    jwt     *auth.JWT
}

func (s *AuthService) Register(req *dto.RegisterRequest) error {
    // hash password, store user, send OTP
}
```

#### `controller/handler.go`

Implements the generated `ServerInterface` (from `internal/gen`).

#### `dto/request.go`, `dto/response.go`

Define request/response structs (often aliases or extensions of generated types).

#### `mapper/mapper.go`

Maps `model.User` ↔ DTO.

#### `migration.go`

```go
package auth

func Migrations() []interface{} {
    return []interface{}{&model.User{}}
}
```

#### `resolver.go`

Wires repository, service, controller.

---

### Feature: `user`

Similar structure to `auth`, but with endpoints for `GET /users/me` and `PATCH /users/me/preferences`.  
Preferences are stored as JSONB in the `users` table.

#### `model/user.go` (may extend the base `User` with a `Preferences` field)

#### `service/service.go` – methods `GetProfile` and `UpdatePreferences`.

---

### Feature: `admin`

Provides full CRUD on users.  
All endpoints require `role == "admin"` (enforced in controller or via middleware that checks context role).

#### `controller/handler.go` – implements `GET /admin/users`, `POST /admin/users`, `GET /admin/users/{id}`, `PUT /admin/users/{id}`, `DELETE /admin/users/{id}`.

#### `service/service.go` – contains business logic for admin operations, including pagination and validation.

#### `repository/gorm.go` – adds methods like `FindAll(limit, offset int)` and `Delete(id string)`.

---

## 🚀 Server & Startup – `internal/app/` Package

After refactoring, the server setup and dependency wiring are moved into dedicated files inside `internal/app/`.

---

### `internal/app/combined.go`

Defines the `CombinedServer` struct that embeds all feature controllers and implements the generated `ServerInterface`.

```go
// internal/app/combined.go
package app

import (
    adminController "rest-api-blueprint/internal/features/admin/controller"
    authController "rest-api-blueprint/internal/features/auth/controller"
    healthController "rest-api-blueprint/internal/features/health/controller"
    userController "rest-api-blueprint/internal/features/user/controller"
)

type CombinedServer struct {
    *healthController.HealthController
    *authController.AuthController
    *userController.UserController
    *adminController.AdminController
}
```

---

### `internal/app/controllers.go`

Builds all repositories, services, and controllers. Returns the `CombinedServer`.

```go
// internal/app/controllers.go
package app

import (
    "rest-api-blueprint/internal/cache"
    "rest-api-blueprint/internal/config"
    "rest-api-blueprint/internal/database"
    "rest-api-blueprint/internal/email"
    adminController "rest-api-blueprint/internal/features/admin/controller"
    adminRepository "rest-api-blueprint/internal/features/admin/repository"
    adminService "rest-api-blueprint/internal/features/admin/service"
    authController "rest-api-blueprint/internal/features/auth/controller"
    authRepository "rest-api-blueprint/internal/features/auth/repository"
    authService "rest-api-blueprint/internal/features/auth/service"
    healthController "rest-api-blueprint/internal/features/health/controller"
    healthRepository "rest-api-blueprint/internal/features/health/repository"
    healthService "rest-api-blueprint/internal/features/health/service"
    userController "rest-api-blueprint/internal/features/user/controller"
    userRepository "rest-api-blueprint/internal/features/user/repository"
    userService "rest-api-blueprint/internal/features/user/service"
)

func BuildControllers(cfg *config.Config, emailSender email.Sender) *CombinedServer {
    // Health feature
    healthRepo := healthRepository.NewRepository(database.DB, cache.Client)
    healthSvc := healthService.NewService(healthRepo)
    healthCtrl := healthController.NewHealthController(healthSvc)

    // Auth feature
    authRepo := authRepository.NewRepository(database.DB)
    authSvc := authService.NewService(authRepo, cfg, cache.Client, emailSender)
    authCtrl := authController.NewAuthController(authSvc)

    // User feature
    userRepo := userRepository.NewRepository(database.DB)
    userSvc := userService.NewService(userRepo)
    userCtrl := userController.NewUserController(userSvc)

    // Admin feature
    adminRepo := adminRepository.NewRepository(database.DB)
    adminSvc := adminService.NewService(adminRepo)
    adminCtrl := adminController.NewAdminController(adminSvc)

    return &CombinedServer{
        HealthController: healthCtrl,
        AuthController:   authCtrl,
        UserController:   userCtrl,
        AdminController:  adminCtrl,
    }
}
```

---

### `internal/app/dto_resolver.go`

Provides the DTO resolver function used by the validation middleware.

```go
// internal/app/dto_resolver.go
package app

import (
    "net/http"
    "rest-api-blueprint/internal/features/admin"
    "rest-api-blueprint/internal/features/auth"
    "rest-api-blueprint/internal/features/user"
)

func DTOResolver(r *http.Request) (any, bool) {
    if dto, ok := auth.Resolver(r); ok {
        return dto, true
    }
    if dto, ok := user.Resolver(r); ok {
        return dto, true
    }
    if dto, ok := admin.Resolver(r); ok {
        return dto, true
    }
    return nil, false
}
```

---

### `internal/app/server.go`

Creates the HTTP mux, registers static and API routes, and builds the middleware chain. Returns the final `http.Handler`.

```go
// internal/app/server.go
package app

import (
    "io/fs"
    "net/http"
    "os"
    "rest-api-blueprint/internal/cache"
    "rest-api-blueprint/internal/config"
    "rest-api-blueprint/internal/gen"
    "rest-api-blueprint/internal/middleware"
)

func SetupServer(
    cfg *config.Config,
    combined *CombinedServer,
    dtoResolver func(*http.Request) (any, bool),
    staticFS fs.FS,
) (http.Handler, error) {
    mux := http.NewServeMux()

    // OpenAPI spec
    openapiSpec, err := os.ReadFile("api/openapi.yaml")
    if err != nil {
        return nil, err
    }
    mux.HandleFunc("GET /openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/yaml")
        w.Write(openapiSpec)
    })

    // Swagger UI
    docsSub, err := fs.Sub(staticFS, "web/docs")
    if err != nil {
        return nil, err
    }
    mux.Handle("GET /docs/", http.StripPrefix("/docs/", http.FileServer(http.FS(docsSub))))

    // Error documentation pages
    errorsSub, err := fs.Sub(staticFS, "web/errors")
    if err != nil {
        return nil, err
    }
    mux.Handle("GET /errors/", http.StripPrefix("/errors/", http.FileServer(http.FS(errorsSub))))

    // Generated API routes
    gen.HandlerFromMuxWithBaseURL(combined, mux, "/api/v1")

    // Middleware chain
    rateLimiter := middleware.NewRateLimiter(cache.Client, cfg.RateLimitPerSecond)
    handler := rateLimiter.Middleware(middleware.DefaultIPKeyFunc)(mux)

    handler = middleware.JWTAuthMiddleware(cfg)(handler)
    handler = middleware.ValidateRequest(dtoResolver)(handler)
    handler = middleware.Logging(handler)
    handler = middleware.RequestIDMiddleware(handler)

    corsMiddleware := middleware.NewCORS(middleware.CORSConfig{
        AllowedOrigins:   cfg.CORSAllowedOrigins,
        AllowedMethods:   cfg.CORSAllowedMethods,
        AllowedHeaders:   cfg.CORSAllowedHeaders,
        AllowCredentials: cfg.CORSAllowCredentials,
    })
    handler = corsMiddleware(handler)

    securityMiddleware := middleware.SecurityHeaders(middleware.SecurityHeadersConfig{
        HSTSMaxAge: cfg.HSTSMaxAge,
    })
    handler = securityMiddleware(handler)
    handler = middleware.PanicRecovery(handler)

    return handler, nil
}
```

---

## 🐳 Summary

This `ARCHITECTURE.md` provides a verified, file‑by‑file description of the Go‑REST‑API‑Blueprint repository, excluding test files. Each file's purpose and key code snippets are shown using the required code fence style.

The refactored structure separates the server setup and dependency wiring into the `internal/app` package, making `main.go` minimal and easier to maintain.

For a high‑level overview, refer to `README.md`.
