# Build stage
FROM golang:1.25.4-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build binary for ARM64 (AWS Graviton)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build \
    -ldflags="-w -s" \
    -o /app/api \
    ./cmd/api

# Runtime stage - ARM64-compatible base image
FROM alpine:3.22

# Install dependencies: ca-certificates for HTTPS, curl for health checks, ffmpeg for video processing
RUN apk --no-cache add ca-certificates curl ffmpeg

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/api .

# Health check (using curl with -f flag for proper GET requests)
HEALTHCHECK --interval=30s --timeout=5s --start-period=60s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./api"]
