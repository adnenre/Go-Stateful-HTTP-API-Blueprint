# =============================================================================
# REST API Blueprint Makefile
# =============================================================================

# Variables
IMAGE_NAME      := rest-api-blueprint
DEV_IMAGE_NAME  := rest-api-blueprint:dev
DOCKER          := docker
DOCKER_COMPOSE  := docker-compose
GO              := go

# Phony targets (no file produced)
.PHONY: help install-tools install-air generate scaffold-feature run dev test clean
.PHONY: docker-up docker-down docker-logs docker-dev docker-build docker-clean

# -----------------------------------------------------------------------------
# Help
# -----------------------------------------------------------------------------
help:
	@echo "Available targets:"
	@echo "  Setup:"
	@echo "    make install-tools      - Install oapi-codegen"
	@echo "    make install-air        - Install air (live reload)"
	@echo ""
	@echo "  Development:"
	@echo "    make generate           - Generate server stubs from openapi.yaml"
	@echo "    make scaffold-feature name=X - Create full layered structure for a new feature"
	@echo "    make run                - Run the server locally (no live reload)"
	@echo "    make dev                - Run with live reload (air) - local only"
	@echo "    make test               - Run all tests"
	@echo "    make clean              - Remove generated files"
	@echo ""
	@echo "  Docker Compose (full stack with PostgreSQL):"
	@echo "    make docker-up          - Start all services (app, postgres) in detached mode"
	@echo "    make docker-down        - Stop and remove all containers"
	@echo "    make docker-logs        - Tail logs from all services"
	@echo "    make docker-dev         - Alias for docker-up (development environment)"
	@echo "    make docker-build       - Rebuild the app image (useful after dependency changes)"
	@echo "    make docker-clean       - Remove containers, volumes, and images"

# -----------------------------------------------------------------------------
# Tool installation
# -----------------------------------------------------------------------------
install-tools:
	@echo "🔧 Installing oapi-codegen..."
	@$(GO) install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
	@echo "✅ Installed. Ensure $$($(GO) env GOPATH)/bin is in your PATH."

install-air:
	@echo "🔧 Installing air (live reload)..."
	@$(GO) install github.com/air-verse/air@latest
	@echo "✅ Installed. Ensure $$($(GO) env GOPATH)/bin is in your PATH."

# -----------------------------------------------------------------------------
# Code generation & scaffolding
# -----------------------------------------------------------------------------
generate:
	@echo "🔧 Generating Go server stubs from api/openapi.yaml (types + std-http)..."
	@mkdir -p internal/gen
	@which oapi-codegen > /dev/null || (echo "❌ oapi-codegen not found. Run: make install-tools" && exit 1)
	@oapi-codegen -generate types,std-http -package gen api/openapi.yaml > internal/gen/api.gen.go
	@test -s internal/gen/api.gen.go || (echo "❌ Generation failed – empty file. Check openapi.yaml syntax." && exit 1)
	@echo "✅ Generated internal/gen/api.gen.go"

scaffold-feature:
	@if [ -z "$(name)" ]; then \
		echo "❌ Usage: make scaffold-feature name=<feature-name>"; \
		exit 1; \
	fi
	@echo "📁 Creating missing directories and files for feature '$(name)'..."
	@mkdir -p internal/features/$(name)/controller \
	           internal/features/$(name)/service \
	           internal/features/$(name)/repository \
	           internal/features/$(name)/model \
	           internal/features/$(name)/mapper \
	           internal/features/$(name)/dto \
	           internal/features/$(name)/tests/unit \
	           internal/features/$(name)/tests/integration
	@$(MAKE) _create_file file=internal/features/$(name)/controller/handler.go content='package controller\n\nimport "net/http"\n\ntype Controller struct {}\n\n// TODO: implement generated ServerInterface methods'
	@$(MAKE) _create_file file=internal/features/$(name)/service/interface.go content='package service\n\ntype Service interface {}'
	@$(MAKE) _create_file file=internal/features/$(name)/service/service.go content='package service\n\ntype service struct{}\n\nfunc NewService() Service {\n\treturn &service{}\n}'
	@$(MAKE) _create_file file=internal/features/$(name)/repository/interface.go content='package repository\n\ntype Repository interface {}'
	@$(MAKE) _create_file file=internal/features/$(name)/repository/gorm.go content='package repository\n\ntype gormRepository struct{}\n\nfunc NewRepository() Repository {\n\treturn &gormRepository{}\n}'
	@$(MAKE) _create_file file=internal/features/$(name)/model/entity.go content='package model\n\n// TODO: define GORM entity'
	@$(MAKE) _create_file file=internal/features/$(name)/mapper/mapper.go content='package mapper\n\n// TODO: convert between model and dto'
	@$(MAKE) _create_file file=internal/features/$(name)/dto/request.go content='package dto\n\n// TODO: request DTOs'
	@$(MAKE) _create_file file=internal/features/$(name)/dto/response.go content='package dto\n\n// TODO: response DTOs'
	@$(MAKE) _create_file file=internal/features/$(name)/tests/unit/handler_test.go content='package unit\n\n// TODO: unit tests for controller'
	@$(MAKE) _create_file file=internal/features/$(name)/tests/integration/api_test.go content='package integration\n\n// TODO: integration tests for the feature'
	@echo "✅ Scaffolded missing parts for feature '$(name)' (no existing files were overwritten)"

# Helper: create file only if it doesn't exist
_create_file:
	@if [ ! -f "$(file)" ]; then \
		echo "$(content)" > "$(file)"; \
	fi

# -----------------------------------------------------------------------------
# Local development (without Docker)
# -----------------------------------------------------------------------------
run: generate
	@echo "🚀 Starting server..."
	@$(GO) run main.go

dev:
	@echo "🔥 Starting development server with live reload (air)..."
	@which air > /dev/null || (echo "❌ air not found. Run: make install-air" && exit 1)
	@air

test:
	@echo "🧪 Running tests..."
	@$(GO) test ./internal/features/.../tests/... -v

clean:
	@echo "🧹 Cleaning up..."
	@rm -f internal/gen/api.gen.go
	@echo "✅ Done"

# -----------------------------------------------------------------------------
# Docker Compose (full environment with PostgreSQL)
# -----------------------------------------------------------------------------
docker-up:
	@echo "🐳 Starting all services with Docker Compose (detached)..."
	@$(DOCKER_COMPOSE) up -d
	@echo "✅ Services running. Access API at http://localhost:$$(grep ^SERVER_PORT .env | cut -d '=' -f2)"
	@echo "   Logs: make docker-logs"

docker-down:
	@echo "🛑 Stopping and removing containers..."
	@$(DOCKER_COMPOSE) down
	@echo "✅ Done"

docker-logs:
	@echo "📋 Tailing logs (Ctrl+C to exit)..."
	@$(DOCKER_COMPOSE) logs -f

docker-dev: docker-up
	@echo "🌟 Development environment ready. API available at http://localhost:$$(grep ^SERVER_PORT .env | cut -d '=' -f2)"
	@echo "   Source code is mounted for hot reload (air)."

docker-build:
	@echo "🐳 Building app image using Docker Compose..."
	@$(DOCKER_COMPOSE) build app
	@echo "✅ Image built. Use 'make docker-up' to start."

docker-clean:
	@echo "🧹 Removing containers, volumes, and images..."
	@$(DOCKER_COMPOSE) down -v --rmi local
	@echo "✅ Cleaned up"