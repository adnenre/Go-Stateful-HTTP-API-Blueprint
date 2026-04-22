#!/bin/bash

set -e

echo "📁 Generating REST API Blueprint folder structure (health feature)..."

# Create directories
mkdir -p api
mkdir -p internal/gen
mkdir -p internal/features/health/controller
mkdir -p internal/features/health/service
mkdir -p internal/features/health/dto
mkdir -p internal/features/health/tests/unit
mkdir -p internal/features/health/tests/integration

# Create empty files with a placeholder comment
cat > api/openapi.yaml << 'EOF'
# OpenAPI contract (to be defined)
EOF

cat > internal/gen/api.gen.go << 'EOF'
// Generated code (will be created by oapi-codegen)
package gen
EOF

cat > internal/features/health/controller/handler.go << 'EOF'
package controller

// TODO: implement generated ServerInterface
EOF

cat > internal/features/health/service/interface.go << 'EOF'
package service

// HealthService defines the business logic interface
type HealthService interface {
	// TODO: add methods
}
EOF

cat > internal/features/health/service/service.go << 'EOF'
package service

// healthService implements HealthService
type healthService struct{}

func NewHealthService() HealthService {
	return &healthService{}
}
EOF

cat > internal/features/health/dto/response.go << 'EOF'
package dto

// HealthResponse DTO (optional, generated types may be used instead)
EOF

cat > internal/features/health/tests/unit/handler_test.go << 'EOF'
package unit

// TODO: unit tests for health controller
EOF

cat > internal/features/health/tests/integration/api_test.go << 'EOF'
package integration

// TODO: integration tests for health API
EOF

cat > main.go << 'EOF'
package main

func main() {
	// TODO: wire features and start server
}
EOF

cat > go.mod << 'EOF'
module rest-api-blueprint

go 1.21
EOF

cat > Makefile << 'EOF'
.PHONY: generate run test

generate:
	oapi-codegen -package gen -generate types,server,spec api/openapi.yaml > internal/gen/api.gen.go

run: generate
	go run main.go

test:
	go test ./... -v
EOF

cat > README.md << 'EOF'
# REST API Blueprint

Contract‑first, feature‑based layered API template (health feature only).
EOF

echo "✅ Done. Folder structure created with empty placeholder files."