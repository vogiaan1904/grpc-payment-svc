FROM golang:1.21-alpine as builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build both binaries
RUN go build -o /app/grpc-server ./cmd/grpc
RUN go build -o /app/http-server ./cmd/http

# Create final image
FROM alpine:latest

WORKDIR /app

# Copy binaries from builder stage
COPY --from=builder /app/grpc-server /app/
COPY --from=builder /app/http-server /app/

# Copy config files if needed
COPY --from=builder /app/.env* /app/ 2>/dev/null || true

# Default command is to run the HTTP server
# You can override with docker-compose or at runtime
CMD ["/app/http-server"] 