# =============================================================================
# REST API Blueprint Makefile
# =============================================================================

# Variables
IMAGE_NAME      := rest-api-blueprint
DEV_IMAGE_NAME  := rest-api-blueprint:dev
DOCKER          := docker
GO              := go

# Phony targets (no file produced)
.PHONY: help install-tools install-air generate scaffold-feature run dev test clean
.PHONY: docker-build docker-build-dev docker-run docker-dev docker-clean

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
	@echo "    make dev                - Run with live reload (air)"
	@echo "    make test               - Run all tests"
	@echo "    make clean              - Remove generated files"
	@echo ""
	@echo "  Docker:"
	@echo "    make docker-build       - Build production Docker image"
	@echo "    make docker-build-dev   - Build development Docker image (live reload)"
	@echo "    make docker-run         - Run production container"
	@echo "    make docker-dev         - Run development container with live reload"
	@echo "    make docker-clean       - Remove Docker images"

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
# Local development
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
# Docker
# -----------------------------------------------------------------------------
docker-build:
	@echo "🐳 Building production Docker image..."
	@$(DOCKER) build -t $(IMAGE_NAME):latest -f Dockerfile .

docker-build-dev:
	@echo "🐳 Building development Docker image (with live reload)..."
	@$(DOCKER) build -t $(DEV_IMAGE_NAME) -f Dockerfile.dev .

docker-run:
	@echo "🐳 Running production container..."
	@$(DOCKER) run -p 8080:8080 $(IMAGE_NAME):latest

docker-dev:
	@echo "🐳 Running development container with live reload..."
	@$(DOCKER) run -p 8080:8080 -v $(PWD):/app $(DEV_IMAGE_NAME)

docker-clean:
	@echo "🧹 Removing Docker images..."
	-@$(DOCKER) rmi $(IMAGE_NAME):latest $(DEV_IMAGE_NAME) 2>/dev/null || true
	@echo "✅ Done"