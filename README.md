# REST API Blueprint

[![Docker Pulls](https://img.shields.io/docker/pulls/adnenre/rest-api-blueprint)](https://hub.docker.com/r/adnenrebai/rest-api-blueprint)

A reusable, contract‑first REST API template built with **pure Go (net/http)** and a **feature‑based layered architecture**.  
Every feature follows the same pattern: `controller → service → repository → model → mapper → dto`.  
The API contract (OpenAPI 3.0) is the single source of truth – all code is generated from it.

## ✨ Key Features

- **Contract‑first** – Define your API in `openapi.yaml`, then generate type‑safe server stubs.
- **No external web framework** – Only the standard library (`net/http`) and a code generator.
- **Feature‑based layered architecture** – Each feature is isolated (controller, service, repository, model, mapper, dto, tests), making it easy to scale or split into microservices later.
- **Built‑in health endpoint** – Enterprise‑ready health check with status, uptime, version, and optional dependency checks.
- **OpenAPI UI** – Swagger documentation embedded in the binary.
- **Makefile** – Automates generation, scaffolding, running, testing, and cleaning.
- **Microservice‑ready** – Designed to be deployed as a monolith today and split into microservices tomorrow with minimal refactoring.

## 🐳 Running with Docker

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
curl http://localhost:8080/v1/health
```

> The Docker image is built and pushed automatically on every tag push (e.g., `v1.0.0`). The `:main` tag is updated on pushes to the `main` branch.

## 📁 Project Structure

```
rest-api-blueprint/
├── api/
│   └── openapi.yaml               # API contract (source of truth)
├── internal/
│   ├── gen/                       # Generated code (types, server interface, router)
│   │   └── api.gen.go
│   └── features/                  # Each feature is a vertical slice
│       └── health/                # Health feature (fully implemented)
│           ├── controller/        # HTTP handlers (implements gen.ServerInterface)
│           ├── service/           # Business logic
│           ├── repository/        # Data access (placeholder)
│           ├── model/             # GORM entity (placeholder)
│           ├── mapper/            # Converts model ↔ dto
│           ├── dto/               # Request/response DTOs
│           └── tests/             # Unit & integration tests
├── main.go                        # Wires all features, starts server
├── go.mod
├── Makefile
└── README.md
```

## 🚀 Getting Started

### Prerequisites

- Go 1.21+
- `make` (optional, but recommended)
- `oapi-codegen` (installed automatically by `make install-tools`)

### Clone and Initialise

```bash
git clone <your-repo> rest-api-blueprint
cd rest-api-blueprint
make install-tools   # installs oapi-codegen
```

### Run the Health Endpoint

```bash
make run
```

Then test:

```bash
curl http://localhost:8080/v1/health
```

Response:

```json
{
  "status": "success",
  "data": {
    "status": "healthy",
    "timestamp": "2026-04-22T12:34:56Z",
    "uptime": "0s",
    "version": "dev"
  }
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

| Command                        | Description                                                              |
| ------------------------------ | ------------------------------------------------------------------------ |
| `make install-tools`           | Installs `oapi-codegen` (once).                                          |
| `make install-air`             | Installs `air` (live reload).                                            |
| `make generate`                | Generates `internal/gen/api.gen.go` from `api/openapi.yaml`.             |
| `make scaffold-feature name=X` | Creates full layered structure for a new feature `X` (never overwrites). |
| `make run`                     | Generates stubs (if needed) and starts the server.                       |
| `make dev`                     | Starts server with live reload (requires `air`).                         |
| `make test`                    | Runs all unit and integration tests.                                     |
| `make clean`                   | Removes generated files.                                                 |

## 🏁 Current Status & Roadmap

The **health feature** is fully implemented and serves as a working example.  
The blueprint is **production‑ready** as a foundation and **microservice‑ready** – you can build new features (auth, scores, leaderboard) using the same pattern.

### What’s already done

- ✅ Contract‑first with OpenAPI 3.0
- ✅ Pure `net/http` server (no external frameworks)
- ✅ Feature‑based layered architecture (controller, service, repository, model, mapper, dto, tests)
- ✅ Code generation via `oapi-codegen`
- ✅ Scaffolding for new features
- ✅ Health endpoint with unit and integration tests
- ✅ Live reload (air) for development
- ✅ Makefile for common tasks
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
