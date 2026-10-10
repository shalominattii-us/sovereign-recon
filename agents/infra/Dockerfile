# Infra Agent - Multi-stage Go Build
# AEGENTIX Sovereign Recon - Infra Agent
# Port: 9090 (whitelisted)
# Network: egress policy restricted

# ─────────────────────────────────────────────────────────────────
# STAGE 1: BUILD
# ─────────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev ca-certificates

# Set working directory
WORKDIR /build

# Copy go mod files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with CGO disabled for static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.version=1.0.0 -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o /infra-agent \
    ./cmd/infra-agent

# ─────────────────────────────────────────────────────────────────
# STAGE 2: RUNTIME (DISTROLESS)
# ─────────────────────────────────────────────────────────────────
FROM gcr.io/distroless/static:nonroot

# Copy binary from builder
COPY --from=builder /infra-agent /infra-agent

# Copy config template
COPY --from=builder /build/configs/config.yaml /etc/infra-agent/config.yaml

# Expose whitelisted port
EXPOSE 9090

# Run as non-root user
USER nonroot:nonroot

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["/infra-agent", "-health"]

# Entry point
ENTRYPOINT ["/infra-agent"]