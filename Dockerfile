# =========================================================================
# Multi-Stage Build Dockerfile for GCP Cloud Run
# =========================================================================

# Stage 1: Build binary
FROM golang:alpine AS builder

WORKDIR /app

# Install TLS certificates and git
RUN apk add --no-cache ca-certificates git

# Cache Go modules layer
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Compile static Linux amd64 binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/blindtag-server ./cmd/server/main.go

# Stage 2: Minimal runtime image
FROM alpine:3.21

WORKDIR /app

# HTTPS certificates and timezone
RUN apk --no-cache add ca-certificates tzdata

# Run as non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# Copy binary from builder
COPY --from=builder /app/blindtag-server /app/blindtag-server

# Cloud Run defaults
EXPOSE 8080
ENV PORT=8080
ENV APP_ENV=production

ENTRYPOINT ["/app/blindtag-server"]
