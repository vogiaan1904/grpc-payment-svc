# Build variables
BINARY_GRPC=payment-grpc-server
BINARY_HTTP=payment-http-server
BUILD_DIR=./build

# Go build flags
GO_BUILD_FLAGS=-ldflags="-s -w"

.PHONY: all build clean run-grpc run-http

all: build

build: build-grpc build-http

build-grpc:
	@echo "Building gRPC server..."
	@mkdir -p $(BUILD_DIR)
	@go build $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_GRPC) ./cmd/grpc

build-http:
	@echo "Building HTTP server..."
	@mkdir -p $(BUILD_DIR)
	@go build $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_HTTP) ./cmd/http

clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)

run-grpc: build-grpc
	@echo "Running gRPC server..."
	@$(BUILD_DIR)/$(BINARY_GRPC)

run-http: build-http
	@echo "Running HTTP server..."
	@$(BUILD_DIR)/$(BINARY_HTTP)

run-all: build
	@echo "Running both servers..."
	@$(BUILD_DIR)/$(BINARY_GRPC) & $(BUILD_DIR)/$(BINARY_HTTP)

docker:
	@echo "Building Docker image..."
	@docker build -t payment-service:latest .

docker-run-http:
	@echo "Running HTTP server in Docker..."
	@docker run -p 8080:8080 --env-file .env payment-service:latest /app/http-server

docker-run-grpc:
	@echo "Running gRPC server in Docker..."
	@docker run -p 50055:50055 --env-file .env payment-service:latest /app/grpc-server 