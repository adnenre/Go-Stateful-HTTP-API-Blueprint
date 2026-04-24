# REST API Blueprint

[![Docker Pulls](https://img.shields.io/docker/pulls/adnenrebai/rest-api-blueprint)](https://hub.docker.com/r/adnenrebai/rest-api-blueprint)
[![CI](https://github.com/adnenre/Go-REST-API-Blueprint/actions/workflows/ci.yml/badge.svg)](https://github.com/adnenre/Go-REST-API-Blueprint/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/adnenre/Go-REST-API-Blueprint)](https://goreportcard.com/report/github.com/adnenre/Go-REST-API-Blueprint)

A reusable, contract‑first REST API template built with **pure Go (net/http)** and a **feature‑based layered architecture**.  
Every feature follows the same pattern: `controller → service → repository → model → mapper → dto`.  
The API contract (OpenAPI 3.0) is the single source of truth – all code is generated from it.

## ✨ Key Features

- **Contract‑first** – Define your API in `openapi.yaml`, then generate type‑safe server stubs.
- **No external web framework** – Only the standard library (`net/http`) and a code generator.
- **Feature‑based layered architecture** – Each feature is isolated (controller, service, repository, model, mapper, dto, tests), making it easy to scale or split into microservices later.
- **Enterprise‑ready health endpoint** – Real checks for PostgreSQL and Redis, returns `200`/`503` with detailed `checks` map.
- **Distributed rate limiting** – Redis‑based token bucket, per client IP, returns `429` with `Retry-After` headers.
- **Request correlation** – `X-Request-Id` header automatically generated, stored in context, and logged.
- **CORS & security headers** – Configurable CORS, plus `X-Content-Type-Options`, `X-Frame-Options`, HSTS, CSP, etc.
- **RFC 7807 error handling** – Standardised `application/problem+json` error responses.
- **Structured JSON logging** – `log/slog` with request ID, method, path, status, latency.
- **Docker Compose** – Full stack (PostgreSQL, Redis, Go app) with hot reload (`air`).
- **GitHub Actions CI/CD** – Tests with service containers, builds and pushes Docker image on tags.
- **OpenAPI UI** – Swagger documentation embedded in the binary.
- **Makefile** – Automates generation, scaffolding, running, testing, Docker management.
- **Microservice‑ready** – Designed to be deployed as a monolith today and split into microservices tomorrow with minimal refactoring.

## ✅ Implemented Enterprise Features (Detailed)

### 1. Project Infrastructure

- [x] Structured configuration (`internal/config`) with `.env` support, fail‑fast validation, no hardcoded secrets.
- [x] Structured JSON logging (`internal/logger`) using `log/slog`.
- [x] Docker Compose stack with PostgreSQL, Redis, and Go app (development with hot reload using `air`).
- [x] `Makefile` targets: `docker-up`, `docker-down`, `docker-logs`, `docker-dev`, `docker-build`, `docker-clean`, `docker-rebuild`.

### 2. Database & Caching

- [x] PostgreSQL connection with GORM, connection pooling (`internal/database`).
- [x] Redis client (`internal/cache`) with health check.

### 3. Health Endpoint (Enterprise‑grade)

- [x] Real database ping (2s timeout) and Redis ping.
- [x] Follows strict layered architecture: `controller → service → repository → model → mapper → dto`.
- [x] Returns `200 OK` if all dependencies are healthy, `503 Service Unavailable` otherwise.
- [x] Includes `checks` map with per‑dependency status (e.g., `database: "ok"`, `redis: "ok"`).
- [x] Unit and integration tests.

### 4. Middleware Pipeline

- [x] **Request ID middleware** – generates/accepts `X-Request-ID` header, stores ID in context.
- [x] **Logging middleware** – logs each request with `request_id`, method, path, status, latency, remote IP.
- [x] **Distributed rate limiting** (Redis‑based) – per client IP, configurable via `RATE_LIMIT_PER_SEC`.
- [x] Rate limiter returns `429 Too Many Requests` with `Retry-After` headers.
- [x] **CORS middleware** – configurable origins, methods, headers, credentials (via environment variables).
- [x] **Security headers middleware** – adds `X-Content-Type-Options`, `X-Frame-Options`, `X-XSS-Protection`, `Strict-Transport-Security` (configurable HSTS max‑age), `Referrer-Policy`, `Content-Security-Policy`, cache control.
- [x] Middleware order: `SecurityHeaders → CORS → RequestID → Logging → RateLimiter`.

### 5. Observability & Correlation

- [x] All logs are JSON (including request logs).
- [x] Request ID correlates logs across a single request.

### 6. Development Experience

- [x] OpenAPI contract (`api/openapi.yaml`) as source of truth.
- [x] Code generation (`oapi-codegen`) for server stubs.
- [x] Scaffolding command (`make scaffold-feature`) for new vertical slices.
- [x] Example health feature fully implemented and tested.

---

## 🐳 Quick Start with Docker

You can run the pre‑built Docker image from Docker Hub:

```bash
docker pull adnenrebai/rest-api-blueprint:main
docker run -p 8080:8080 adnenrebai/rest-api-blueprint:main
```

Or use a specific version:

```bash
docker pull adnenrebai/rest-api-blueprint:v1.0.0
docker run -p 8080:8080 adnenrebai/rest-api-blueprint:v1.0.0
```

Then test the health endpoint:

```bash
curl http://localhost:8080/api/v1/health
```

Example response:

```bash
{
  "status": "success",
  "data": {
    "status": "healthy",
    "timestamp": "2026-04-24T10:00:00Z",
    "uptime": "1m2s",
    "version": "1.0.0",
    "checks": {
      "database": "ok",
      "redis": "ok"
    }
  }
}
```

> The Docker image is built and pushed automatically on every tag push (e.g., `v1.0.0`). The `:main` tag is updated on pushes to the `main` branch.

## 📁 Project Structure

## 📁 Project Structure

````
rest-api-blueprint/
├── api/
│   └── openapi.yaml                    # API contract (source of truth)
├── internal/
│   ├── gen/                            # Generated code (types, server interface)
│   │   └── api.gen.go
│   ├── config/                         # Configuration loading (.env + env vars)
│   ├── logger/                         # Structured JSON logging (slog)
│   ├── database/                       # GORM connection & connection pool
│   ├── cache/                          # Redis client
│   ├── errors/                         # RFC 7807 error handling (domain errors, problem details)
│   ├── middleware/                     # Security, CORS, RequestID, Logging, RateLimiter
│   └── features/                       # Vertical slices
│       └── health/                     # Health feature (fully implemented)
│           ├── controller/             # HTTP handlers (implements gen.ServerInterface)
│           ├── service/                # Business logic (calls repository)
│           ├── repository/             # Data access (real DB/Redis ping)
│           ├── model/                  # GORM entity (optional)
│           ├── mapper/                 # Model ↔ DTO conversion
│           ├── dto/                    # Request/response DTOs
│           └── tests/                  # Unit & integration tests
├── .github/
│   └── workflows/                      # CI/CD pipelines (ci.yml, cd.yml)
├── docker-compose.yml                  # PostgreSQL, Redis, and Go app with hot reload
├── .env.example                        # Template for environment variables
├── main.go                             # Wires all features, starts server with middleware
├── go.mod
├── Makefile
└── README.md
--

## 🚀 Getting Started

### Prerequisites

- Go 1.26+
- `make` (optional, but recommended)
- `oapi-codegen` (installed automatically by `make install-tools`)

### Clone and Initialise

```bash
git clone <your-repo> rest-api-blueprint
cd rest-api-blueprint
make install-tools   # installs oapi-codegen
````

### Run the Health Endpoint

```bash
make run
```

Then test:

```bash
curl http://localhost:8080/api/v1/health
```

Response:

```json
{
  "data": {
    "checks": {
      "database": "ok",
      "redis": "ok"
    },
    "status": "healthy",
    "timestamp": "2026-04-24T15:26:41.782319008Z",
    "uptime": "26m28s",
    "version": "1.0.0"
  },
  "status": "success"
}
```

Swagger UI is available at [http://localhost:8080/swagger/](http://localhost:8080/swagger/).

## 🧱 Adding a New Feature (e.g., `auth`)

The workflow is **contract‑first** – always start with the OpenAPI specification.

### Step 1: Add Endpoints to `api/openapi.yaml`

Add your new paths and schemas. Example for a login endpoint:

```yaml
paths:
  /v1/auth/login:
    post:
      summary: Login user
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                username:
                  type: string
                password:
                  type: string
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/LoginResponse"
```

### Step 2: Generate Server Stubs

```bash
make generate
```

This updates `internal/gen/api.gen.go` with:

- New Go structs for request/response DTOs.
- New methods in `ServerInterface` (e.g., `PostV1AuthLogin`).

### Step 3: Scaffold the Feature Folder

```bash
make scaffold-feature name=auth
```

This creates the full layered structure for `auth`:

```
internal/features/auth/
├── controller/handler.go
├── service/interface.go
├── service/service.go
├── repository/interface.go
├── repository/gorm.go
├── model/entity.go
├── mapper/mapper.go
├── dto/request.go
├── dto/response.go
└── tests/
    ├── unit/handler_test.go
    └── integration/api_test.go
```

### Step 4: Implement the Layers

1. **Define the model** – `internal/features/auth/model/entity.go` (GORM entity).
2. **Implement the repository** – `repository/gorm.go` with database operations.
3. **Write the service** – `service/service.go` (business logic).
4. **Create the mapper** – `mapper/mapper.go` to convert between model and DTO.
5. **Implement the controller** – `controller/handler.go` (satisfies `gen.ServerInterface`).

Example controller stub:

```go
package controller

import (
    "net/http"
    "rest-api-blueprint/internal/features/auth/service"
    "rest-api-blueprint/internal/gen"
)

type AuthController struct {
    svc service.Service
}

func NewAuthController(svc service.Service) *AuthController {
    return &AuthController{svc: svc}
}

func (c *AuthController) PostV1AuthLogin(w http.ResponseWriter, r *http.Request) {
    // Parse request, call service, map response using mapper
}
```

### Step 5: Wire the New Controller in `main.go`

```go
// Inside main()
authRepo := repository.NewRepository()
authSvc := service.NewService(authRepo)
authCtrl := controller.NewAuthController(authSvc)
gen.HandlerFromMux(authCtrl, mux)
```

### Step 6: Run and Test

```bash
make run
curl -X POST http://localhost:8080/v1/auth/login -d '{"username":"alice","password":"pass"}' -H "Content-Type: application/json"
```

## 🧪 Testing

- **Unit tests** – `internal/features/*/tests/unit/` (mock service/repository).
- **Integration tests** – `internal/features/*/tests/integration/` (use a real database or test HTTP server).

Run all tests:

```bash
make test
```

## 🛠️ Makefile Commands

## 🛠️ Makefile Targets

| Command                        | Description                                                      |
| ------------------------------ | ---------------------------------------------------------------- |
| `make install-tools`           | Installs `oapi-codegen` (required for generation).               |
| `make install-air`             | Installs `air` (live reload).                                    |
| `make generate`                | Regenerates `internal/gen/api.gen.go` from `openapi.yaml`.       |
| `make scaffold-feature name=X` | Creates full layered structure for a new feature `X`.            |
| `make run`                     | Runs the server locally (no live reload).                        |
| `make dev`                     | Runs with live reload (`air`).                                   |
| `make test`                    | Runs all unit and integration tests (requires PostgreSQL/Redis). |
| `make clean`                   | Removes generated files.                                         |
| `make docker-up`               | Starts services in detached mode.                                |
| `make docker-down`             | Stops containers.                                                |
| `make docker-logs`             | Tails logs from all services.                                    |
| `make docker-dev`              | Starts stack with logs attached (press Ctrl+C to stop).          |
| `make docker-build`            | Rebuilds the app image.                                          |
| `make docker-clean`            | Removes containers, volumes, images, and build cache.            |
| `make docker-rebuild`          | Full clean rebuild (runs `docker-clean` then `docker-dev`).      |

---

## 🏁 Current Status & Roadmap

The **health feature** is fully implemented and serves as a working example.  
The blueprint is **production‑ready** as a foundation and **microservice‑ready** – you can build new features (auth, scores, leaderboard) using the same pattern.

### What’s already done

- ✅ Contract‑first with OpenAPI 3.0
- ✅ Pure `net/http` server (no external frameworks)
- ✅ Feature‑based layered architecture (controller, service, repository, model, mapper, dto, tests)
- ✅ Code generation via `oapi-codegen`
- ✅ Scaffolding for new features
- ✅ Health endpoint with unit and integration tests (real PostgreSQL + Redis pings, 200/503 with `checks` map)
- ✅ Live reload (`air`) for development
- ✅ Makefile for common tasks (including Docker Compose targets)
- ✅ Structured configuration (`.env` + env vars, fail‑fast validation)
- ✅ Structured JSON logging (`log/slog` with request ID)
- ✅ PostgreSQL connection (GORM, connection pooling)
- ✅ Redis client (used for rate limiting and health checks)
- ✅ Distributed rate limiting (Redis‑based, per IP, returns 429)
- ✅ Request ID middleware (`X-Request-Id` header, context, logs)
- ✅ CORS middleware (configurable via env)
- ✅ Security headers middleware (XSS, clickjacking, HSTS, CSP, cache control)
- ✅ RFC 7807 error handling (`application/problem+json`)
- ✅ Docker Compose stack (PostgreSQL, Redis, Go app with hot reload)
- ✅ GitHub Actions CI (tests with PostgreSQL/Redis service containers)
- ✅ GitHub Actions CD (builds and pushes Docker image on tags)
- ✅ README with clear instructions

### What you can build next

- **Auth** – user registration, login, JWT cookies
- **Admin** – user management with RBAC

### When to split into microservices

The architecture supports splitting without major refactoring. Each feature is isolated, uses its own database schema (or separate database), and communicates via HTTP. When the monolith grows, you can extract a feature into a standalone service by:

- Copying the feature folder and `common/` package
- Adding a standalone `main.go`
- Routing traffic via an API gateway

## 📄 License

MIT

## Author

- github: https://github.com/adnenre
- website: https://adnenre.dev
