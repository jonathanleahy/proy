# Multi-stage build for optimal image size
# Stage 1: Build the Go binary
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o proxy cmd/proxy/main.go

# Stage 2: Create the final lightweight image
FROM alpine:latest

# Install certificates for HTTPS support
RUN apk --no-cache add ca-certificates

# Create non-root user for security
RUN addgroup -g 1000 proxy && \
    adduser -D -u 1000 -G proxy proxy

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/proxy .

# Copy web assets
COPY --from=builder /build/web ./web

# Create recordings directory
RUN mkdir -p recordings && \
    chown -R proxy:proxy /app

# Switch to non-root user
USER proxy

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://0.0.0.0:8080/health || exit 1

# Default environment variables
ENV PROXY_PORT=8080 \
    PROXY_HOST=0.0.0.0 \
    PROXY_RECORDINGS_DIR=/app/recordings \
    PROXY_MODE=playback \
    PROXY_TLS_SKIP_VERIFY=true

# Volume for recordings
VOLUME ["/app/recordings"]

# Run the proxy
ENTRYPOINT ["./proxy"]