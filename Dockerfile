# ==============================================================================
# VECDIM Core Backend API & Spatial Ingestion Platform
# Multi-stage Docker build optimized for Koyeb, Render, and Cloud Container Runtimes
# ==============================================================================

# Stage 1: Build static Go binary
FROM golang:1.24-bookworm AS builder

WORKDIR /app

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy application source code
COPY . .

# Compile optimized static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o vecdim-api ./cmd/api/main.go

# Stage 2: Minimal, secure production runtime
FROM debian:bookworm-slim

WORKDIR /app

# Install CA certificates and curl for container health probes
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Copy binary and all necessary data/analytics assets
COPY --from=builder /app/vecdim-api /app/vecdim-api
COPY --from=builder /app/data /app/data
COPY --from=builder /app/api /app/api
COPY --from=builder /app/R /app/R
COPY --from=builder /app/outputs /app/outputs

# Default runtime environment
ENV PORT=8080
ENV APP_ENV=production

EXPOSE 8080

# Health check probe for Koyeb / Kubernetes
HEALTHCHECK --interval=15s --timeout=5s --retries=3 \
  CMD curl -f http://localhost:${PORT}/api/v1/health || exit 1

# Execute server
CMD ["/app/vecdim-api"]
