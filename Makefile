.PHONY: proto clean build run docker-dev

SERVICE_NAME := payment-svc
PROTO_DIR := ../protos/proto
PROTOGEN_DIR := protogen

proto:
	@echo "Generating protobuf..."
	protoc --proto_path=$(PROTO_DIR) \
		--go_out=$(PROTOGEN_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(PROTOGEN_DIR) --go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/payment.proto

clean:
	@echo "Cleaning..."
	rm -rf $(PROTOGEN_DIR)/golang/payment/*.pb.go

build:
	@echo "Building..."
	go build -o bin/server cmd/server/main.go

run:
	@echo "Running..."
	go run cmd/server/main.go

docker-dev:
	@echo "Starting docker compose development environment..."
	docker-compose -f docker-compose.dev.yml up -d 