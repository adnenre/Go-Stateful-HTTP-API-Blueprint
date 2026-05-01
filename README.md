# Go Stateful HTTP API Blueprint v3.2.0

[![Docker Pulls](https://img.shields.io/docker/pulls/adnenrebai/rest-api-blueprint)](https://hub.docker.com/r/adnenrebai/rest-api-blueprint)
[![CI](https://github.com/adnenre/Go-REST-API-Blueprint/actions/workflows/ci.yml/badge.svg)](https://github.com/adnenre/Go-REST-API-Blueprint/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/adnenre/Go-REST-API-Blueprint)](https://goreportcard.com/report/github.com/adnenre/Go-REST-API-Blueprint)

> **⚠️ Terminology note:** This project is a **stateful HTTP API**, not a REST API anymore. It intentionally uses server‑side state (Redis) for refresh token rotation, rate limiting, OTP storage, and session endpoints – all of which violate the statelessness constraint of REST. If you need a truly RESTful API, you would have to remove these features.

A reusable, **contract‑first Stateful HTTP API template** built with **pure Go (net/http)** and a **feature‑based layered architecture**.  
Every feature follows the same pattern: `controller → service → repository → model → mapper → dto`.  
The API contract (OpenAPI 3.0) is the single source of truth – all code is generated from it.

## ✨ Key Features

- **Contract‑first** – Define your API in `openapi.yaml`, then generate type‑safe server stubs.
- **No external web framework** – Only the standard library (`net/http`) and a code generator.
- **Feature‑based layered architecture** – Each feature is isolated (controller, service, repository, model, mapper, dto, tests), making it easy to scale or split into microservices later.
- **Enterprise‑ready health endpoint** – Real checks for PostgreSQL and Redis, returns `200`/`503` with detailed `checks` map.
- **Dual‑mode authentication** – Supports both `httpOnly` cookies (secure web BFF) and `Authorization: Bearer` headers (mobile apps) from the same endpoints.
- **Refresh token rotation** – One‑time use refresh tokens stored in Redis, automatically rotated on each refresh request.
- **Session endpoint** – `GET /api/v1/auth/session` returns current user without needing a separate profile request.
- **JWT authentication with OTP verification** – Register, login, email‑based OTP activation, and password reset flows.
- **User profile & preferences** – `GET /users/me` and `PATCH /users/me/preferences`.
- **Admin user management** – Full CRUD on users (`/admin/users`) with role‑based access (admin only).
- **Distributed rate limiting** – Redis‑based token bucket, per client IP, returns `429` with `Retry-After` headers.
- **Request correlation** – `X-Request-Id` header automatically generated, stored in context, and logged.
- **CORS & security headers** – Configurable CORS, plus `X-Content-Type-Options`, `X-Frame-Options`, HSTS, CSP, etc.
- **RFC 7807 error handling** – Standardised `application/problem+json` error responses.
- **Global request validation** – Automatic DTO validation with field‑specific RFC 7807 errors.
- **Structured JSON logging** – `log/slog` with request ID, method, path, status, latency.
- **Docker Compose** – Full stack (PostgreSQL, Redis, Go app) with hot reload (`air`).
- **GitHub Actions CI/CD** – Tests with service containers, builds and pushes Docker image on tags.
- **OpenAPI UI** – Swagger documentation embedded in the binary.
- **Makefile** – Automates generation, scaffolding, running, testing, Docker management.
- **Microservice‑ready** – Designed to be deployed as a monolith today and split into microservices tomorrow with minimal refactoring.

## ❓ Why “Go Stateful HTTP API” and not “REST”?

REST (Representational State Transfer) requires **statelessness** – the server must not store any client context between requests. This project deliberately stores:

- Refresh tokens in Redis (with rotation state)
- Rate limiting counters per IP in Redis
- OTP codes and password reset tokens in Redis (with TTL)
- Session information (via the `/session` endpoint)

These features are **incompatible with REST**. The API is therefore a **stateful HTTP API**, sometimes called a “REST‑inspired” or “resource‑oriented HTTP API”. It remains a well‑structured, production‑ready design – just not RESTful.

If you require full REST compliance, you would need to:

- Remove refresh token rotation and use self‑contained JWTs only (no server‑side token storage)
- Replace Redis rate limiting with a stateless algorithm (e.g., client‑driven)
- Eliminate the session endpoint and store any session data in self‑contained tokens
- Move OTP and reset tokens to self‑contained signed tokens

For the vast majority of applications, the stateful approach is simpler, more secure, and perfectly acceptable. We choose honesty over marketing.

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

### 4. Authentication & User Management

- [x] **JWT utility** – generate/validate tokens (`internal/auth/jwt.go`).
- [x] **Dual‑mode JWT middleware** – accepts token from `access_token` httpOnly cookie (for web) OR `Authorization: Bearer` header (for mobile).
- [x] **User registration** – `POST /api/v1/auth/register` (email, username, password, optional avatar) – returns `202 Accepted`, stores user as `pending`.
- [x] **OTP verification** – `POST /api/v1/auth/verify-otp` (6‑digit OTP stored in Redis, 10 min TTL) activates account and returns `access_token` + `refresh_token` (JSON body for mobile) and also sets `httpOnly` cookies (for web).
- [x] **User login** – `POST /api/v1/auth/login` returns tokens in JSON body and sets `httpOnly` cookies.
- [x] **Refresh token rotation** – `POST /api/v1/auth/refresh` accepts refresh token (cookie for web, JSON body for mobile), rotates token pair, stores new tokens in Redis, sets new cookies, returns JSON tokens.
- [x] **Session endpoint** – `GET /api/v1/auth/session` returns current user profile using the access token (cookie or Bearer header).
- [x] **User profile** – `GET /api/v1/users/me` (protected, returns `firstName`, `lastName`, `emailVerified` along with other fields).
- [x] **User preferences** – `PATCH /api/v1/users/me/preferences` (stored in separate table).
- [x] **Admin CRUD** – full user management under `/api/v1/admin/users` (list, create, get by ID, update, delete) – only accessible with `admin` role.
- [x] **Password hashing** – bcrypt.
- [x] **Role‑based access control** – `admin` vs `user` (checked in admin endpoints).
- [x] **Password reset** – `POST /api/v1/auth/password-reset/request` (sends reset token via email, always 202) and `POST /api/v1/auth/password-reset/confirm` (updates password with token).
- [x] **Email sender** – mock (logs to stdout) for development, async SMTP sender with worker pool for production.

### 5. Middleware Pipeline

- [x] **Request ID middleware** – generates/accepts `X-Request-ID` header, stores ID in context.
- [x] **Logging middleware** – logs each request with `request_id`, method, path, status, latency, remote IP.
- [x] **Distributed rate limiting** (Redis‑based) – per client IP, configurable via `RATE_LIMIT_PER_SEC`.
- [x] Rate limiter returns `429 Too Many Requests` with `Retry-After` headers.
- [x] **CORS middleware** – configurable origins, methods, headers, credentials (via environment variables).
- [x] **Security headers middleware** – adds `X-Content-Type-Options`, `X-Frame-Options`, `X-XSS-Protection`, `Strict-Transport-Security` (configurable HSTS max‑age), `Referrer-Policy`, `Content-Security-Policy`, cache control.
- [x] Middleware order: `SecurityHeaders → CORS → RequestID → JWTAuth → Logging → RateLimiter`.

### 6. Request Validation & Error Documentation

- [x] **Global validation middleware** – uses per‑feature resolvers to validate request bodies, restores body, returns `422` with field‑specific errors.
- [x] **RFC 7807 absolute error types** – error `type` URIs point to static HTML documentation pages (e.g., `/errors/validation.html`).
- [x] **Swagger UI** – interactive API documentation served at `/docs/`, consuming `/openapi.yaml`.
- [x] **Static error pages** – human‑readable explanations for each error type served under `/errors/`.
- [x] **Panic recovery middleware** – catches panics and returns `InternalError` with a proper type.
- [x] Public documentation paths (`/docs/`, `/errors/`, `/openapi.yaml`, `/favicon.ico`) exempt from JWT and rate limiting.

### 7. Observability & Correlation

- [x] All logs are JSON (including request logs).
- [x] Request ID correlates logs across a single request.

### 8. Development Experience

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
docker pull adnenrebai/rest-api-blueprint:v3.2.0
docker run -p 8080:8080 adnenrebai/rest-api-blueprint:v3.2.0
```

Then test the health endpoint:

```bash
curl http://localhost:8080/api/v1/health
```

Example response:

```
{
  "status": "success",
  "data": {
    "status": "healthy",
    "timestamp": "2026-04-30T10:00:00Z",
    "uptime": "1m2s",
    "version": "1.0.0",
    "checks": {
      "database": "ok",
      "redis": "ok"
    }
  }
}
```

> The Docker image is built and pushed automatically on every tag push (e.g., `v3.2.0`). The `:main` tag is updated on pushes to the `main` branch.

### Interactive API Documentation

Open `http://localhost:8080/docs/` in your browser to explore the API using Swagger UI. The OpenAPI specification is available at `http://localhost:8080/openapi.yaml`.

### Error Documentation

Every RFC 7807 error response contains a `type` URI (e.g., `/errors/validation.html`). These URIs resolve to human‑readable HTML pages explaining the error. You can also browse them at `http://localhost:8080/errors/`.

### 🔐 OTP Verification Flow

After registration, the user receives an OTP via email (mock sender logs to stdout). The account must be activated with:

```bash
curl -X POST http://localhost:8080/api/v1/auth/verify-otp \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","otp":"123456"}'
```

Only then the user can log in.

### 🔄 Refresh Token Flow (Web vs Mobile)

- **Web (BFF)**: The refresh token is stored in an `httpOnly` cookie. Call `POST /api/v1/auth/refresh` without body – the cookie is automatically sent.
- **Mobile app**: Send the refresh token in the JSON body: `{ "refresh_token": "..." }`. The backend will return new tokens in the JSON body.

Example for web (using cookie):

```bash
curl -X POST -b cookies.txt -c cookies.txt http://localhost:8080/api/v1/auth/refresh
```

Example for mobile (using JSON body):

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"<token>"}'
```

### 📧 Password Reset Flow

Request a reset token (always returns 202, no user enumeration):

```bash
curl -X POST http://localhost:8080/api/v1/auth/password-reset/request \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com"}'
```

Confirm with the token:

```bash
curl -X POST http://localhost:8080/api/v1/auth/password-reset/confirm \
  -H "Content-Type: application/json" \
  -d '{"token":"<token>","new_password":"NewPass123"}'
```

### 🔐 Creating an Admin User

To test admin endpoints, you need a user with the `admin` role. By default, registration creates users with role `user`. Promote a user to admin using the database (while the stack is running):

```bash
# Register a user first (or use an existing one)
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin123","username":"admin"}'

# Verify OTP (copy from logs) then promote role
docker exec -it rest_api_postgres psql -U postgres -d rest_api_blueprint \
  -c "UPDATE users SET role = 'admin' WHERE email = 'admin@example.com';"
```

Now you can log in as `admin@example.com` and obtain an admin token to call the admin endpoints.

## 📁 Project Structure

```
go-stateful-http-api/
├── api/
│   └── openapi.yaml      # API contract (source of truth)
├── internal/
│   ├── gen/              # Generated code (types, server interface)
│   │   └── api.gen.go
│   ├── config/           # Configuration loading (.env + env vars)
│   ├── logger/           # Structured JSON logging (slog)
│   ├── database/         # GORM connection & connection pool
│   ├── cache/            # Redis client
│   ├── auth/             # JWT utility (shared)
│   ├── errors/           # RFC 7807 error handling (domain errors, problem details)
│   ├── middleware/       # Security, CORS, RequestID, JWTAuth, Logging, RateLimiter, Validation, PanicRecovery
│   └── features/         # Vertical slices
│       ├── health/       # Health endpoint (implemented)
│       ├── auth/         # User registration, login, OTP, password reset, refresh, session (implemented)
│       ├── user/         # Profile & preferences (implemented)
│       └── admin/        # Admin CRUD on users (implemented)
├── .github/
│   └── workflows/        # CI/CD pipelines (ci.yml, cd.yml)
├── web/                  # Embedded static files (Swagger UI, error pages)
├── docker-compose.yml    # PostgreSQL, Redis, and Go app with hot reload
├── .env.example          # Template for environment variables
├── main.go               # Wires all features, starts server with middleware
├── go.mod
├── Makefile
└── README.md
```

## 🚀 Getting Started

### Prerequisites

- Go 1.26+
- `make` (optional, but recommended)
- `oapi-codegen` (installed automatically by `make install-tools`)

### Clone and Initialise

```bash
git clone <your-repo> go-stateful-http-api
cd go-stateful-http-api
make install-tools   # installs oapi-codegen
```

### Run the Health Endpoint

```bash
make run
```

Then test:

```bash
curl http://localhost:8080/api/v1/health
```

Response:

```
{
  "data": {
    "checks": {
      "database": "ok",
      "redis": "ok"
    },
    "status": "healthy",
    "timestamp": "2026-04-30T15:26:41.782319008Z",
    "uptime": "26m28s",
    "version": "1.0.0"
  },
  "status": "success"
}
```

Swagger UI is available at [http://localhost:8080/docs/](http://localhost:8080/docs/).

### Test Authentication & OTP Flow

```bash
# Register (returns 202, OTP sent to email)
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"pass123","username":"testuser"}'

# Check logs for OTP (e.g., 654321)

# Verify OTP to activate account and get tokens
curl -X POST http://localhost:8080/api/v1/auth/verify-otp \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","otp":"654321"}'

# Then login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"pass123"}'
```

### Test Validation Error (missing username)

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"strongpass"}'
```

Expected response: `422 Unprocessable Entity` with field‑specific error for `username`.

## 🧱 Adding a New Feature (example: `products`)

The workflow is **contract‑first** – always start with the OpenAPI specification.

### Step 1: Add Endpoints to `api/openapi.yaml`

Add your new paths and schemas. Example for a product catalog:

```yaml
paths:
  /v1/products:
    get:
      summary: List all products
      operationId: listProducts
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: "#/components/schemas/Product"
    post:
      summary: Create a new product
      operationId: createProduct
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/NewProduct"
      responses:
        "201":
          description: Created
  /v1/products/{id}:
    get:
      summary: Get a product by ID
      operationId: getProduct
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        "200":
          description: OK
    put:
      summary: Update a product
      operationId: updateProduct
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/UpdateProduct"
      responses:
        "200":
          description: Updated
    delete:
      summary: Delete a product
      operationId: deleteProduct
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        "204":
          description: Deleted

components:
  schemas:
    Product:
      type: object
      properties:
        id:
          type: string
        name:
          type: string
        price:
          type: number
    NewProduct:
      type: object
      required: [name, price]
      properties:
        name:
          type: string
        price:
          type: number
    UpdateProduct:
      type: object
      properties:
        name:
          type: string
        price:
          type: number
```

### Step 2: Generate Server Stubs

```bash
make generate
```

This updates `internal/gen/api.gen.go` with:

- New Go structs for request/response DTOs.
- New methods in `ServerInterface` (e.g., `ListProducts`, `CreateProduct`, etc.).

### Step 3: Scaffold the Feature Folder

```bash
make scaffold-feature name=products
```

This creates the full layered structure for `products`:

```
internal/features/products/
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

1. **Define the model** – `internal/features/products/model/entity.go` (GORM entity).
2. **Implement the repository** – `repository/gorm.go` with database operations (CRUD).
3. **Write the service** – `service/service.go` (business logic).
4. **Create the mapper** – `mapper/mapper.go` to convert between model and DTO.
5. **Implement the controller** – `controller/handler.go` (satisfies `gen.ServerInterface`).

Example controller stub:

```go
package controller

import (
	"net/http"
	"go-stateful-http-api/internal/features/products/service"
	"go-stateful-http-api/internal/gen"
)

type ProductsController struct {
	svc service.Service
}

func NewProductsController(svc service.Service) *ProductsController {
	return &ProductsController{svc: svc}
}

func (c *ProductsController) ListProducts(w http.ResponseWriter, r *http.Request) {
	products, err := c.svc.List(r.Context())
	if err != nil {
		// use errors.WriteProblemSimple or domain error
		return
	}
	// map and respond
}
```

### Step 5: Wire the New Controller in `main.go`

```go
// Inside main()
productsRepo := repository.NewRepository(database.DB)
productsSvc := service.NewService(productsRepo)
productsCtrl := controller.NewProductsController(productsSvc)
// Add to combined server struct
```

### Step 6: Run and Test

```bash
make run
curl http://localhost:8080/v1/products
curl -X POST http://localhost:8080/v1/products -d '{"name":"Laptop","price":999}' -H "Content-Type: application/json"
```

## 🧪 Testing

- **Unit tests** – `internal/features/*/tests/unit/` (mock service/repository).
- **Integration tests** – `internal/features/*/tests/integration/` (use a real database or test HTTP server).

Run all tests:

```bash
make test
```

## 🛠️ Makefile Commands

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

The **health, auth (with OTP, password reset, refresh token rotation, and session), user, and admin features** are fully implemented and serve as working examples.  
The blueprint is **production‑ready** as a foundation and **microservice‑ready** – you can build new features (products, scores, leaderboard, etc.) using the same pattern.

### What’s already done

- ✅ Contract‑first with OpenAPI 3.0
- ✅ Pure `net/http` server (no external frameworks)
- ✅ Feature‑based layered architecture (controller, service, repository, model, mapper, dto, tests)
- ✅ Code generation via `oapi-codegen`
- ✅ Scaffolding for new features
- ✅ Health endpoint (`/api/v1/health`) with unit and integration tests (real PostgreSQL + Redis pings, 200/503 with `checks` map)
- ✅ **Dual‑mode authentication** (cookies for web, Bearer for mobile)
- ✅ **Refresh token rotation** with Redis storage
- ✅ **Session endpoint** (`/api/v1/auth/session`)
- ✅ **JWT authentication with OTP verification**: `POST /auth/register` (202, pending), `POST /auth/verify-otp` (activates, returns tokens), `POST /auth/login` (tokens)
- ✅ **Password reset**: `POST /auth/password-reset/request` and `POST /auth/password-reset/confirm`
- ✅ **User profile**: `GET /users/me` (protected, returns extended profile fields)
- ✅ **User preferences**: `PATCH /users/me/preferences` (store/update preferences)
- ✅ **Admin user management**: Full CRUD on `/admin/users` (list, create, get by ID, update, delete) – only accessible with `admin` role
- ✅ JWT authentication middleware (skips public paths, injects claims into context; reads token from cookie or Bearer header)
- ✅ Live reload (`air`) for development
- ✅ Makefile for common tasks (including Docker Compose targets)
- ✅ Structured configuration (`.env` + env vars, fail‑fast validation)
- ✅ Structured JSON logging (`log/slog` with request ID)
- ✅ PostgreSQL connection (GORM, connection pooling)
- ✅ Redis client (used for rate limiting, health checks, OTP, reset tokens, refresh token storage)
- ✅ Distributed rate limiting (Redis‑based, per IP, returns 429)
- ✅ Request ID middleware (`X-Request-Id` header, context, logs)
- ✅ CORS middleware (configurable via env)
- ✅ Security headers middleware (XSS, clickjacking, HSTS, CSP, cache control)
- ✅ RFC 7807 error handling (`application/problem+json`)
- ✅ **Global request validation middleware** (DTO validation, field‑specific errors)
- ✅ **Absolute error documentation URIs** (point to static HTML pages)
- ✅ **Swagger UI** (`/docs/`) and static error pages (`/errors/`)
- ✅ **Panic recovery middleware** (returns structured internal error)
- ✅ **Asynchronous email sender** (mock for development, SMTP with worker pool for production)
- ✅ **Redis‑based OTP and password reset tokens** (with TTL)
- ✅ Docker Compose stack (PostgreSQL, Redis, Go app with hot reload)
- ✅ GitHub Actions CI (tests with PostgreSQL/Redis service containers)
- ✅ GitHub Actions CD (builds and pushes Docker image on tags)
- ✅ README with clear instructions

### What you can build next

- **Pagination, filtering, sorting** – for list endpoints.
- **Prometheus metrics** – monitor API performance.
- **OpenTelemetry tracing** – distributed tracing with Jaeger.
- **Webhooks** – event notifications to external services.

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
