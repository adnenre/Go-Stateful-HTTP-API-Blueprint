FROM golang:1.26.2 AS builder

WORKDIR /app

# Install oapi-codegen (code generator)
RUN go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and generate OpenAPI stubs
COPY . .
RUN make generate

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Final stage
FROM alpine:3.22
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]