.PHONY: generate run test

generate:
	oapi-codegen -package gen -generate types,server,spec api/openapi.yaml > internal/gen/api.gen.go

run: generate
	go run main.go

test:
	go test ./... -v
