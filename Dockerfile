# ============================================
# Dockerfile — slimify
# Multi-stage build for minimal image size
# ============================================

# --- Build stage ---
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src

# Copy go module files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X github.com/NotHarshhaa/slimify/cmd.Version=$(git describe --tags --always --dirty 2>/dev/null || echo dev) -X github.com/NotHarshhaa/slimify/cmd.Commit=$(git rev-parse --short HEAD 2>/dev/null || echo none)" \
    -o /slimify .

# --- Production stage ---
FROM alpine:3.20

RUN apk add --no-cache ca-certificates

COPY --from=builder /slimify /usr/local/bin/slimify

ENTRYPOINT ["slimify"]
