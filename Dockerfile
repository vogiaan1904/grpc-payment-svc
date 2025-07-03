# Build stage
FROM golang:1.23.6-alpine AS builder

# Install ca-certificates (no need for git anymore)
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and pre-generated proto files
COPY . .

# Build both gRPC and HTTP servers
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o grpc-server ./cmd/grpc

RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o http-server ./cmd/http

# Production stage
FROM alpine:latest AS production

# Install ca-certificates for SSL/TLS connections
RUN apk --no-cache add ca-certificates

# Create non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -S appuser -u 1001 -G appgroup

WORKDIR /app

# Copy binaries from builder stage
COPY --from=builder /app/grpc-server .
COPY --from=builder /app/http-server .

# Copy config files
COPY --from=builder /app/config ./config

# Copy proto files (if needed at runtime)
COPY --from=builder /app/protogen ./protogen

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app
USER appuser

# Expose both gRPC and HTTP ports
EXPOSE 50056 8080

# Default to gRPC server (can be overridden in docker-compose)
CMD ["./grpc-server"] 