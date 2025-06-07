# Build variables
BINARY_GRPC=payment-grpc-server
BINARY_HTTP=payment-http-server
BUILD_DIR=./build

# Go build flags
GO_BUILD_FLAGS=-ldflags="-s -w"
APP_NAME=payment-svc

.PHONY: all build clean run-grpc run-http

run-server:
	@echo "Starting $(APP_NAME)..."
	go run cmd/grpc/main.go

run-http:
	@echo "Starting $(APP_NAME) HTTP server..."
	go run cmd/http/main.go

protoc-all:
	$(MAKE) protoc PAYMENT_PROTO=protos/proto/payment.proto OUT_DIR=protogen/golang/payment
	$(MAKE) protoc PAYMENT_PROTO=protos/proto/order.proto OUT_DIR=protogen/golang/order
	$(MAKE) protoc PAYMENT_PROTO=protos/proto/product.proto OUT_DIR=protogen/golang/product

protoc:
	protoc --go_out=$(OUT_DIR) --go_opt=paths=source_relative \
	--go-grpc_out=$(OUT_DIR) --go-grpc_opt=paths=source_relative \
	-I=protos/proto $(PAYMENT_PROTO)